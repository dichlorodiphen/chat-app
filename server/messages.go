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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Representation of a message in the database and over the wire.
type Message struct {
	ID      string    `bson:"_id,omitempty" json:"id"`
	Author  string    `bson:"author" json:"author"`
	Content string    `bson:"content" json:"content"`
	Votes   int       `json:"votes"`
	Created time.Time `bson:"created" json:"created"`
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
		Author:  r.Header.Get("username"),
		Content: body.Content,
		Votes:   0,
		Created: time.Now(),
	}
	insertResult, err := messagesCollection.InsertOne(s.ctx, message)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast message on websocket.
	message.ID = insertResult.InsertedID.(primitive.ObjectID).Hex()
	serialized, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.hub.broadcast <- serialized

	w.WriteHeader(http.StatusNoContent)
}

// Body of request to the update message endpoint.
type UpdateMessageRequestBody struct {
	Upvoted   bool `json:"upvoted"`
	Downvoted bool `json:"downvoted"`
}

// Endpoint for updating the vote count of a message.
func handleUpdateMessage(s *Server, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	log.Printf("Updating message with ID: %v\n", id)

	// Deserialize request.
	var body UpdateMessageRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Upvoted && body.Downvoted {
		http.Error(w, "upvoted and downvoted cannot both be true", http.StatusBadRequest)
		return
	}

	// Update vote.
	var err error
	username := r.Header.Get("username")
	if body.Upvoted {
		err = s.addUpvote(username, id)
	} else {
		err = s.removeUpvote(username, id)
	}
	if err == nil {
		if body.Downvoted {
			err = s.addDownvote(username, id)
		} else {
			err = s.removeDownvote(username, id)
		}
	}
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Printf("successfully updated message %v with upvoted %v and downvoted %v\n", id, body.Upvoted, body.Downvoted)
}

// Upvote a message (idempotent operation).
func (s Server) addUpvote(username string, id string) error {
	return s.executeAsTransaction(func() error {
		user, err := s.getUser(username)
		if err != nil {
			return err
		}

		if _, ok := user.Upvoted[id]; ok {
			return nil
		}
		user.Upvoted[id] = struct{}{}

		if err := s.updateMessageVotes(id, 1); err != nil {
			return err
		}
		if err := s.updateUserVotes(user); err != nil {
			return err
		}

		return nil
	})
}

// Remove an upvote from a message (idempotent operation).
func (s Server) removeUpvote(username string, id string) error {
	return s.executeAsTransaction(func() error {
		user, err := s.getUser(username)
		if err != nil {
			return err
		}

		if _, ok := user.Upvoted[id]; !ok {
			return nil
		}
		delete(user.Upvoted, id)

		if err := s.updateMessageVotes(id, -1); err != nil {
			return err
		}
		if err := s.updateUserVotes(user); err != nil {
			return err
		}

		return nil
	})
}

// Downvote a message (idempotent operation).
func (s Server) addDownvote(username string, id string) error {
	return s.executeAsTransaction(func() error {
		user, err := s.getUser(username)
		if err != nil {
			return err
		}

		if _, ok := user.Downvoted[id]; ok {
			return nil
		}
		user.Downvoted[id] = struct{}{}

		if err := s.updateMessageVotes(id, -1); err != nil {
			return err
		}
		if err := s.updateUserVotes(user); err != nil {
			return err
		}

		return nil
	})
}

// Remove a downvote from a message (idempotent operation).
func (s Server) removeDownvote(username string, id string) error {
	return s.executeAsTransaction(func() error {
		user, err := s.getUser(username)
		if err != nil {
			return err
		}

		if _, ok := user.Downvoted[id]; !ok {
			return nil
		}
		delete(user.Downvoted, id)

		if err := s.updateMessageVotes(id, 1); err != nil {
			return err
		}
		if err := s.updateUserVotes(user); err != nil {
			return err
		}

		return nil
	})
}

func (s Server) getUser(username string) (User, error) {
	var user User
	if err := s.users.FindOne(s.ctx, bson.M{"username": username}).Decode(&user); err != nil {
		log.Println(err)
		return User{}, err
	}

	return user, nil
}

// Updates the vote count of the given message by n in the database.
func (s Server) updateMessageVotes(id string, n int) error {
	objectID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", objectID}}
	update := bson.M{"$inc": bson.M{"votes": n}}
	if err := s.messages.FindOneAndUpdate(s.ctx, filter, update).Err(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Updates the voted and downvoted messages for the given user in the database.
func (s Server) updateUserVotes(user User) error {
	objectID, _ := primitive.ObjectIDFromHex(user.ID)
	filter := bson.D{{"_id", objectID}}
	update := bson.M{
		"$set": bson.M{
			"upvoted":   user.Upvoted,
			"downvoted": user.Downvoted,
		},
	}
	if err := s.users.FindOneAndUpdate(s.ctx, filter, update).Err(); err != nil {
		log.Println("error updating user")
		return err
	}

	return nil
}

// Execute the given database code as a transaction.
func (s Server) executeAsTransaction(f func() error) error {
	session, err := s.dbClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(s.ctx)

	// TODO: If this will be executed in multiple goroutines, protect w/ mutex.
	// See: https://www.mongodb.com/docs/drivers/go/current/fundamentals/transactions/.
	_, err = session.WithTransaction(s.ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		return nil, f()
	})

	return err
}
