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
	// Client used to connect to the database.
	db *mongo.Client

	// Hub encapsulating websocket connections to server.
	hub *Hub

	// Session manager for storing client-side session data.
	store *sessions.CookieStore

	// Multiplexer for handling routing.
	router *mux.Router
}

func newServer() *Server {
	return &Server{
		// TODO: create db context
		db:  connectToDatabase(),
		hub: newHub(),
		// TODO: Use an actual secret lol.
		store:  sessions.NewCookieStore([]byte("secret")),
		router: mux.NewRouter(),
	}
}
func (s Server) setUpRoutes() {
	s.router.PathPrefix("/").HandlerFunc(s.wrapHandler(echo))
	s.router.Path("/users/signup").
		Methods("POST").
		HandlerFunc(handleSignup)
	s.router.Path("/users/login").
		Methods("POST").
		HandlerFunc(handleLogin)
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(s.hub, w, r)
	})
}

func (s Server) start() {
	defer func() {
		log.Println("Disconnecting MongoDB client.")
		if err := s.db.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("Starting server.")
	go s.hub.run()
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", s.router))
}

// Creates a connection to the database and returns the corresponding Client.
func connectToDatabase() *mongo.Client {
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
	client, err := mongo.Connect(context.TODO(), options)
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
	if r.URL.Path == "/" {
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
