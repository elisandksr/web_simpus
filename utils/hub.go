package utils

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	Clients    map[string]*Client // Map UserID -> Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	mu         sync.Mutex
}

type Message struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if current, ok := h.Clients[client.UserID]; ok && current == client {
				delete(h.Clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			client, ok := h.Clients[message.UserID]
			if ok {
				select {
				case client.Send <- []byte(message.Content):
				default:
					close(client.Send)
					delete(h.Clients, client.UserID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Global Hub
var NotificationHub = NewHub()

func init() {
	go NotificationHub.Run()
}

