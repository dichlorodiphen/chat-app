package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Delay between heartbeats in seconds.
	heartbeatDelay = 25 * time.Second

	// Time before client is considered unresponsive.
	heartbeatTimeout = 30 * time.Second

	// Time before a write is considered failed.
	writeTimeout = 10 * time.Second

	// Time allowed for user to send credentials for authentication setup.
	authTimeout = 10 * time.Second

	// Maximum message size allowed from client.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a client connection to a user on the frontend.
type Client struct {
	// The hub responsible for the client.
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Continuously reads messages from the websocket.
func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.WriteMessage(websocket.CloseMessage, nil)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	// Handle heartbeats.
	c.conn.SetReadDeadline(time.Now().Add(heartbeatTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(heartbeatTimeout))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		c.hub.broadcast <- message
	}
}

// Continuously writes messages from the send queue to the websocket.
func (c *Client) write() {
	ticker := time.NewTicker(heartbeatDelay)
	defer func() {
		ticker.Stop()
		c.conn.WriteMessage(websocket.CloseMessage, nil)
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// If hub closed channel, close connection.
			if !ok {
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println(err)
				return
			}
			w.Write(message)

			// Add any remaining messages from queue.
			for len(c.send) > 0 {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				log.Println(err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		}
	}

}

// Returns a non-nil error if a non-authenticated user tries to establish a websocket connection.
func (c *Client) ensureAuthenticated() error {
	log.Println("Waiting for authentication message from client.")
	c.conn.SetReadDeadline(time.Now().Add(authTimeout))
	_, signedString, err := c.conn.ReadMessage()
	if err != nil {
		log.Println("Did not receive valid credentials before timeout.")
		return err
	}
	log.Printf("Got the following JWT, attempting to verify: %v\n", signedString)
	_, err = verifyJWTToken(string(signedString))

	return err

}

// Handles the creation of a Client when receiving an incoming websocket connection.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming websocket connection from %v\n", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 16),
	}

	if err := client.ensureAuthenticated(); err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	hub.register <- client

	// Start reading from and writing to websocket.
	log.Println("Now reading from and writing to websocket.")
	go client.write()
	go client.read()
}
