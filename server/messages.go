// Routes for create, read, and update operations on messages.
package main

import (
	"encoding/json"
	"fmt"
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
	Vote int `json:"vote"`
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

	// Update vote count.
	var err error
	username := r.Header.Get("username")
	switch body.Vote {
	case 1:
		err = s.upvote(username, id)
	case -1:
		err = s.downvote(username, id)
	default:
		log.Println("bad value for vote")
		http.Error(w, "bad value for vote", http.StatusBadRequest)
		return
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
}

// Upvote a message.
func (s Server) upvote(username string, id string) error {
	session, err := s.dbClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(s.ctx)

	messagesCollection := s.db.Collection("messages")
	usersCollection := s.db.Collection("users")

	_, err = session.WithTransaction(s.ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		changeInVoteCount := 0

		// Get user.
		var user User
		if err := usersCollection.FindOne(s.ctx, bson.M{"username": username}).Decode(&user); err != nil {
			log.Println(err)
			return nil, err
		}

		// Add upvote if not already upvoted.
		if _, ok := user.Upvoted[id]; ok {
			return nil, fmt.Errorf("user %v has already upvoted message %v", username, id)
		}
		user.Upvoted[id] = struct{}{}
		changeInVoteCount++

		// Remove downvote if it exists.
		if _, ok := user.Downvoted[id]; ok {
			delete(user.Downvoted, id)
			changeInVoteCount++
		}

		// Increment vote count on message.
		objectID, _ := primitive.ObjectIDFromHex(id)
		filter := bson.D{{"_id", objectID}}
		update := bson.M{"$inc": bson.M{"votes": changeInVoteCount}}
		if err := messagesCollection.FindOneAndUpdate(s.ctx, filter, update).Err(); err != nil {
			log.Println(err)
			return nil, err
		}

		// Update user.
		objectID, _ = primitive.ObjectIDFromHex(user.ID)
		filter = bson.D{{"_id", objectID}}
		update = bson.M{
			"$set": bson.M{
				"upvoted":   user.Upvoted,
				"downvoted": user.Downvoted,
			},
		}
		if err := usersCollection.FindOneAndUpdate(s.ctx, filter, update).Err(); err != nil {
			log.Println("error updating user")
			return nil, err
		}

		return nil, nil
	})

	return err
}

// Downvote a message.
func (s Server) downvote(username string, id string) error {
	session, err := s.dbClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(s.ctx)

	messagesCollection := s.db.Collection("messages")
	usersCollection := s.db.Collection("users")

	_, err = session.WithTransaction(s.ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		changeInVoteCount := 0

		// Get user.
		var user User
		if err := usersCollection.FindOne(s.ctx, bson.M{"username": username}).Decode(&user); err != nil {
			log.Println(err)
			return nil, err
		}

		// Add downvote if not already downvoted.
		if _, ok := user.Downvoted[id]; ok {
			return nil, fmt.Errorf("user %v has already downvoted message %v", username, id)
		}
		user.Downvoted[id] = struct{}{}
		changeInVoteCount--

		// Remove upvote if it exists.
		if _, ok := user.Upvoted[id]; ok {
			delete(user.Upvoted, id)
			changeInVoteCount--
		}

		// Increment vote count on message.
		objectID, _ := primitive.ObjectIDFromHex(id)
		filter := bson.D{{"_id", objectID}}
		update := bson.M{"$inc": bson.M{"votes": changeInVoteCount}}
		if err := messagesCollection.FindOneAndUpdate(s.ctx, filter, update).Err(); err != nil {
			log.Println(err)
			return nil, err
		}

		// Update user.
		objectID, _ = primitive.ObjectIDFromHex(user.ID)
		filter = bson.D{{"_id", objectID}}
		update = bson.M{
			"$set": bson.M{
				"upvoted":   user.Upvoted,
				"downvoted": user.Downvoted,
			},
		}
		if err := usersCollection.FindOneAndUpdate(s.ctx, filter, update).Err(); err != nil {
			log.Println("error updating user")
			return nil, err
		}

		return nil, nil
	})

	return err
}
