// Routes for user registration and authentication.
package main

import (
	"encoding/json"
	"net/http"
)

type SignupRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement.

	var body SignupRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement.
	return
}
