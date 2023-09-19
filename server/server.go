// The backend that is started in main.go.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

	// The users collection in the database.
	users *mongo.Collection

	// The messages collection in the database.
	messages *mongo.Collection

	// The context for the database connection.
	ctx context.Context

	// Hub encapsulating websocket connections to server.
	hub *Hub

	// Multiplexer for handling routing.
	router *mux.Router
}

// Create a new server.
func newServer() *Server {
	ctx := context.TODO()
	client := connectToDatabase(ctx)
	db := client.Database("admin")

	return &Server{
		dbClient: client,
		db:       db,
		users:    db.Collection("users"),
		messages: db.Collection("messages"),
		ctx:      ctx,
		hub:      newHub(),
		router:   mux.NewRouter(),
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
	messagesRouter := s.router.NewRoute().Subrouter()
	messagesRouter.Use(authenticationMiddleware)
	messagesRouter.Path("/messages").
		Methods("GET", "OPTIONS").
		HandlerFunc(s.wrapHandler(handleGetAllMessages))
	messagesRouter.Path("/messages").
		Methods("POST", "OPTIONS").
		HandlerFunc(s.wrapHandler(handleCreateMessage))
	messagesRouter.Path("/messages/{id}").
		Methods("PATCH", "OPTIONS").
		HandlerFunc(s.wrapHandler(handleUpdateMessage))

	// Websocket for real-time chat.
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		serveWs(s.hub, w, r)
	})
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
		log.Fatal("mongoDB username environment variable not defined.")
	}
	if password == "" {
		log.Fatal("mongoDB password environment variable not defined.")
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
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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
