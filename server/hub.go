package main

// Hub maintains the set of active websocket connections.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from clients.
	broadcast chan []byte

	// Register requests from clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// If client send channel is full, unregister the client.
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}
