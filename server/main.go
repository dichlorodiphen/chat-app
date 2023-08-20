// A minimal echo server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Student struct {
	FirstName string `bson:"first_name,omitempty"`
	LastName  string `bson:"last_name,omitempty"`
	Age       int    `bson:"omitempty"`
}

func main() {
	client := connectToDatabase()
	defer func() {
		log.Println("Disconnecting MongoDB client.")
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	// TODO: Remove db testing code below.
	student := Student{
		FirstName: "Arthur",
		LastName:  "Evans",
		Age:       12,
	}
	collection := client.Database("admin").Collection("students")
	log.Println("Created collection.")
	// TODO: Create context.
	filter := Student{FirstName: "Arthur"}
	var result Student
	log.Println("Performing initial read.")
	if err := collection.FindOne(context.TODO(), filter).Decode(&result); err != nil {
		log.Println("Could not decode result from initial read.")
		log.Println(err)
	} else {
		log.Printf("Result from first read: %v\n", result)
	}
	if _, err := collection.InsertOne(context.TODO(), student); err != nil {
		log.Println(err)
	} else {
		log.Println("Successfully inserted student into collection.")
	}
	if err := collection.FindOne(context.TODO(), filter).Decode(&result); err != nil {
		log.Println("Could not decode result from second read.")
		log.Println(err)
	} else {
		log.Printf("Result from second read: %v\n", result)
	}

	// END: db testing code.

	hub := newHub()
	go hub.run()

	http.HandleFunc("/", echo)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fmt.Println("Starting server.")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
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
