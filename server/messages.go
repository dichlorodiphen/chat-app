// Routes for create, read, and update operations on messages.
package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Endpoint for getting all messages in chat.
func handleGetAllMessages(s *Server, w http.ResponseWriter, r *http.Request) {
	return
}

// Endpoint for creating a new message.
func handleCreateMessage(s *Server, w http.ResponseWriter, r *http.Request) {
	return
}

// Endpoint for updating the vote count of a message.
func handleUpdateMessage(s *Server, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	return
}
