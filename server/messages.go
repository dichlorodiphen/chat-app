// Routes for create, read, and update operations on messages.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Endpoint for getting all messages in chat.
func handleGetAllMessages(s *Server, w http.ResponseWriter, r *http.Request) {
	log.Println("Getting all messages in chat.")
	fmt.Fprintln(w, "So if this were implemented, there would be many messages here. Yeet.")
}

// Endpoint for creating a new message.
func handleCreateMessage(s *Server, w http.ResponseWriter, r *http.Request) {
	log.Println("Creating a new message.")
}

// Endpoint for updating the vote count of a message.
func handleUpdateMessage(s *Server, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	log.Printf("Updating message with ID: %v\n", id)
}
