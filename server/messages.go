// Routes for create, read, and update operations on messages.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// Representation of a message in the database and over the wire.
type Message struct {
	ID      string    `bson:"_id,omitempty" json:"id"`
	Author  string    `bson:"author,omitempty" json:"author"`
	Content string    `bson:"content,omitempty" json:"content"`
	Votes   int       `json:"votes"`
	Created time.Time `bson:"created,omitempty" json:"created"`
}

// Endpoint for getting all messages in chat.
func handleGetAllMessages(s *Server, w http.ResponseWriter, r *http.Request) {
	log.Println("Getting all messages in chat.")

	messagesCollection := s.db.Collection("messages")
	cursor, err := messagesCollection.Find(s.ctx, bson.D{})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	messages := []Message{}
	if err := cursor.All(s.ctx, &messages); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Created.Before(messages[j].Created)
	})

	serialized, err := json.Marshal(messages)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Set Content-Type = application/json?
	if _, err := w.Write(serialized); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Body of request to the create message endpoint.
type CreateMessageRequestBody struct {
	Content string `json:"content"`
}

// Endpoint for creating a new message.
func handleCreateMessage(s *Server, w http.ResponseWriter, r *http.Request) {
	log.Println("Creating a new message.")

	// Deserialize request.
	var body CreateMessageRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add message to database.
	messagesCollection := s.db.Collection("messages")
	message := Message{
		// TODO: maybe add nil check for username header.
		Author:  w.Header().Get("username"),
		Content: body.Content,
		Votes:   0,
		Created: time.Now(),
	}
	if _, err := messagesCollection.InsertOne(s.ctx, message); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Endpoint for updating the vote count of a message.
func handleUpdateMessage(s *Server, w http.ResponseWriter, r *http.Request) {
	// TODO: Use findoneandupdate
	// https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#Collection.FindOneAndUpdate
	id := mux.Vars(r)["id"]
	log.Printf("Updating message with ID: %v\n", id)
}
