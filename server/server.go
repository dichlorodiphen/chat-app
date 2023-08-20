package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	// The connection to the MongoDB database.
	dbClient *mongo.Client

	// The database used for this application.
	db *mongo.Database

	// The context for the database connection.
	ctx context.Context

	// Hub encapsulating websocket connections to server.
	hub *Hub

	// Session manager for storing client-side session data.
	store *sessions.CookieStore

	// Multiplexer for handling routing.
	router *mux.Router
}

func newServer() *Server {
	ctx := context.TODO()
	client := connectToDatabase(ctx)

	return &Server{
		// TODO: create db context
		dbClient: client,
		db:       client.Database("admin"),
		ctx:      ctx,
		hub:      newHub(),
		// TODO: Use an actual secret lol.
		store:  sessions.NewCookieStore([]byte("secret")),
		router: mux.NewRouter(),
	}
}
func (s Server) setUpRoutes() {
	// Users API.
	s.router.Path("/users/signup").
		Methods("POST").
		HandlerFunc(s.wrapHandler(handleSignup))
	s.router.Path("/users/login").
		Methods("POST").
		HandlerFunc(s.wrapHandler(handleLogin))

	// Messsages API.
	s.router.Path("/messages").
		Methods("GET").
		HandlerFunc(s.wrapHandler(handleGetAllMessages))
	s.router.Path("/messages").
		Methods("POST").
		HandlerFunc(s.wrapHandler(handleCreateMessage))
	s.router.Path("/messages/{id}").
		Methods("PATCH").
		HandlerFunc(s.wrapHandler(handleUpdateMessage))

	// Websocket for real-time chat.
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		serveWs(s.hub, w, r)
	})

	// Ping endpoint (for testing).
	s.router.PathPrefix("/ping").HandlerFunc(s.wrapHandler(echo))

}

func (s Server) start() {
	defer func() {
		log.Println("Disconnecting MongoDB client.")
		if err := s.dbClient.Disconnect(s.ctx); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("Starting server.")
	go s.hub.run()
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", s.router))
}

// Creates a connection to the database and returns the corresponding Client.
func connectToDatabase(ctx context.Context) *mongo.Client {
	username := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	password := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
	if username == "" {
		log.Fatal("MongoDB username environment variable not set.")
	}
	if password == "" {
		log.Fatal("MongoDB password environment variable not set.")
	}
	credentials := options.Credential{
		Username: username,
		Password: password,
	}
	options := options.Client().ApplyURI("mongodb://db-service:27017/admin").SetAuth(credentials)
	client, err := mongo.Connect(ctx, options)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("MongoDB client successfully connected.")

	return client
}

func (s *Server) wrapHandler(handler func(s *Server, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		handler(s, w, r)
	}
}

// Echoes back the request path to the client.
func echo(s *Server, w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, "test-session")
	if r.URL.Path == "/ping" {
		session.Values["test"] = "success"
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Println("Stored data in session.")
	} else {
		log.Println("Reading data in session.")
		log.Println(session.Values["test"])
	}

	fmt.Println("Got connection")
	fmt.Fprintf(w, "Path: %q\n", r.URL.Path)
}
