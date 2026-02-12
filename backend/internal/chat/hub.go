package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients by project ID
	clients map[int64]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mutex sync.Mutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if h.clients[client.projectID] == nil {
				h.clients[client.projectID] = make(map[*Client]bool)
			}
			h.clients[client.projectID][client] = true
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if clients, ok := h.clients[client.projectID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
				}
				if len(h.clients[client.projectID]) == 0 {
					delete(h.clients, client.projectID)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// Broadcast sends a message to all clients in a project
func (h *Hub) Broadcast(projectID int64, message []byte) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.clients[projectID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				// Client's send buffer is full, close the connection
				close(client.send)
				delete(clients, client)
			}
		}
	}
}

// GetClientCount returns the number of connected clients for a project
func (h *Hub) GetClientCount(projectID int64) int {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.clients[projectID]; ok {
		return len(clients)
	}
	return 0
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	projectID int64
}

// Message represents a chat message
type Message struct {
	Type       string `json:"type"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

// ServeWs handles WebSocket requests from clients
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	// Get project ID from URL
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		log.Println("Invalid project ID:", err)
		conn.Close()
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		projectID: projectID,
	}

	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Broadcast the message to all clients in the same project
		messageJSON, _ := json.Marshal(message)
		c.hub.Broadcast(c.projectID, messageJSON)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		message, ok := <-c.send
		if !ok {
			// Hub closed the channel
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		w.Write(message)

		// Add queued messages to the current WebSocket message
		n := len(c.send)
		for i := 0; i < n; i++ {
			w.Write([]byte{'\n'})
			w.Write(<-c.send)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}