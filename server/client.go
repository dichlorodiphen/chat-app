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
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	// Handle heartbeats.
	c.conn.SetReadDeadline(time.Now().Add(heartbeatTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(heartbeatDelay))
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
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// If hub closed channel, close connection.
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
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
	hub.register <- client

	// Start reading from and writing to websocket.
	go client.write()
	go client.read()
}
