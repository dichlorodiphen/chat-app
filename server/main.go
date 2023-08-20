// A minimal echo server.
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

func main() {
	client := connectToDatabase()
	defer func() {
		log.Println("Disconnecting MongoDB client.")
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	hub := newHub()
	go hub.run()

	router := mux.NewRouter()
	router.PathPrefix("/").HandlerFunc(echo)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fmt.Println("Starting server.")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", router))
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

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// Echoes back the request path to the client.
func echo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	fmt.Println("Got connection")
	fmt.Fprintf(w, "Path: %q\n", r.URL.Path)
}
