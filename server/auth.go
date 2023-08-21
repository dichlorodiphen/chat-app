// A collection of utility functions for authentication.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Replace this with an environment variable.
var JWT_SIGNING_KEY = []byte("secret")

// Generate a signed JWT token for the given username.
func generateJWTToken(username string) (string, error) {
	claims := JwtClaims{
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(JWT_SIGNING_KEY)
}

// Verifies and extracts claims from a signed JWT token.
func verifyJWTToken(signedString string) (jwt.Claims, error) {
	token, err := jwt.Parse(signedString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return JWT_SIGNING_KEY, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token was invalid")
	}

	return token.Claims, err
}

// Authenticates with JWT token and updates header with claim information.
func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and proceess signed string.
		signedString := r.Header.Get("Authorization")
		if len(signedString) == 0 {
			http.Error(w, "Missing authorization header.", http.StatusUnauthorized)
			return
		}
		signedString = strings.Replace(signedString, "Bearer ", "", 1)

		// Verify signed string and extract claims.
		claims, err := verifyJWTToken(signedString)
		if err != nil {
			log.Println("Error verifying JWT token: " + err.Error())
			http.Error(w, "Error verifying JWT token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		log.Println("Verified JWT token")

		// Update headers with information from claims.
		username := claims.(jwt.MapClaims)["username"].(string)
		r.Header.Set("username", username)

		log.Println("successfully passed through authentication middleware")

		next.ServeHTTP(w, r)
	})
}
