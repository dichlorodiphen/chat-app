// Routes for user registration and authentication.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Number of iterations bcrypt will use to hash the password.
const BCRYPT_ITERATIONS = 12

// Custom JWT claims so that we can extract the username of the user.
type JwtClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Representation of user in database.
type User struct {
	ID       string `bson:"_id,omitempty"`
	Username string `bson:"username"`
	Password []byte `bson:"password"`

	// Set of upvoted messages (by IDs).
	Upvoted map[string]struct{} `bson:"upvoted"`

	// Set of downvoted messages (by IDs).
	Downvoted map[string]struct{} `bson:"downvoted"`
}

// Body of requests to the signup and login endpoints.
type AuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Endpoint for user signup.
func handleSignup(s *Server, w http.ResponseWriter, r *http.Request) {
	// Deserialize request.
	var body AuthRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Username: %v\n", body.Username)
	log.Printf("Password: %v\n", body.Password)
	log.Printf("Body: %v\n", body)

	// TODO: Add validation logic.

	// Check if user already exists.
	alreadyExists, err := s.userExists(body.Username)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if alreadyExists {
		log.Println("Account already exists.")
		http.Error(w, "Account already exists.", http.StatusBadRequest)
		return
	}

	// Add new user to database.
	users := s.db.Collection("users")
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), BCRYPT_ITERATIONS)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newUser := User{
		Username:  body.Username,
		Password:  hash,
		Upvoted:   map[string]struct{}{},
		Downvoted: map[string]struct{}{},
	}
	if _, err := users.InsertOne(s.ctx, newUser); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Compute JWT.
	// TODO: Add in rollback logic on error.
	token, err := generateJWT(body.Username)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, token)
	log.Printf("Created user with username: %v and password %v\n", body.Username, body.Password)
}

// Return whether or not a user exists in our database.
func (s Server) userExists(username string) (bool, error) {
	users := s.db.Collection("users")
	err := users.FindOne(s.ctx, bson.M{"username": username}).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Endpoint for user authentication.
func handleLogin(s *Server, w http.ResponseWriter, r *http.Request) {
	// Deserialize request.
	var body AuthRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user from database.
	users := s.db.Collection("users")
	var user User
	if err := users.FindOne(s.ctx, bson.M{"username": body.Username}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "No account with given username and password.", http.StatusForbidden)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Compare passwords.
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(body.Password)); err != nil {
		http.Error(w, "No account with given username and password.", http.StatusForbidden)
		return
	}

	// Generate JWT.
	token, err := generateJWT(body.Username)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, token)
	log.Printf("Successfully authenticated user with username: %v and password %v\n", body.Username, body.Password)
}
