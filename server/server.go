// The backend that is started in main.go.
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

// The server struct encapsulates the entire state of the backend, including
// both the websocket used for real-time chat and the REST API for the control
// plane.
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

// Create a new server.
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

// Set up the routes in our API.
func (s Server) setUpRoutes() {
	// CORS
	s.router.Use(corsMiddleware)

	// Users API.
	s.router.Path("/users/signup").
		Methods("POST").
		HandlerFunc(s.wrapHandler(handleSignup))
	s.router.Path("/users/login").
		Methods("POST").
		HandlerFunc(s.wrapHandler(handleLogin))

	// Messsages API.
	messagesRouter := s.router.PathPrefix("/messages").Subrouter()
	messagesRouter.Use(authenticationMiddleware)
	messagesRouter.Path("").
		Methods("GET", "OPTIONS").
		HandlerFunc(s.wrapHandler(handleGetAllMessages))
	messagesRouter.Path("").
		Methods("POST", "OPTIONS").
		HandlerFunc(s.wrapHandler(handleCreateMessage))
	messagesRouter.Path("{id}").
		Methods("PATCH", "OPTIONS").
		HandlerFunc(s.wrapHandler(handleUpdateMessage))

	// Websocket for real-time chat.
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		serveWs(s.hub, w, r)
	})

	// Ping endpoint (for testing).
	s.router.PathPrefix("/ping").HandlerFunc(s.wrapHandler(echo))

}

// Begin serving the routes associated with the server's mux.
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		// Don't pass down chain if preflight request.
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) wrapHandler(handler func(s *Server, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
