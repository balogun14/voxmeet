package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (customize in production)
	},
}

const (
	signalExchange = "voxmeet.signal"
)

// Hub manages a set of active WebSocket clients.
type Hub struct {
	clients      map[*Client]bool
	clientsByID  map[string]*Client // userID → Client
	Register     chan *Client
	Unregister   chan *Client
	mu           sync.RWMutex
	rmqPublisher func(exchange, routingKey string, body []byte)
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		clientsByID: make(map[string]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
	}
}

// SetPublisher sets the RabbitMQ publishing function for signal messages.
func (h *Hub) SetPublisher(pub func(exchange, routingKey string, body []byte)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rmqPublisher = pub
}

// Run starts the hub's event loop. Must be called as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.clientsByID[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if h.clientsByID[client.UserID] == client {
					delete(h.clientsByID, client.UserID)
				}
				close(client.Send)
			}
			h.mu.Unlock()
		}
	}
}

// RouteToClient sends a message to a specific client by user ID.
func (h *Hub) RouteToClient(userID string, msg []byte) {
	h.mu.RLock()
	client, ok := h.clientsByID[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.Send <- msg:
		default:
		}
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ConsumeSignalMessages starts consuming from RabbitMQ and routing to clients.
// Runs until ctx is cancelled. Must be called as a goroutine.
func (h *Hub) ConsumeSignalMessages(ctx context.Context, msgs <-chan amqp091.Delivery) {
	for d := range msgs {
		var envelope struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(d.Body, &envelope); err == nil {
			h.RouteToClient(envelope.UserID, d.Body)
		}
		d.Ack(false)
	}
}

// Client represents a single WebSocket connection.
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID string
	mu     sync.Mutex
}

// NewClient creates a new Client.
func NewClient(userID string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}
}

// WritePump sends messages from the Send channel to the WebSocket connection.
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// ReadPump reads messages from the WebSocket connection.
func (c *Client) ReadPump(handler MessageHandler) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(msg, &incoming); err != nil {
			c.sendError("invalid_message", "invalid message format")
			continue
		}

		handler(c, incoming)
	}
}

// sendError sends an error message to the client.
func (c *Client) sendError(code, message string) {
	msg, _ := json.Marshal(map[string]string{
		"type":    "error",
		"code":    code,
		"message": message,
	})
	select {
	case c.Send <- msg:
	default:
	}
}

// IncomingMessage is the base structure for messages from clients.
type IncomingMessage struct {
	Type   string          `json:"type"`
	RoomID string          `json:"room_id,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// MessageHandler is a function that processes an incoming message from a client.
type MessageHandler func(client *Client, msg IncomingMessage)

// Handler returns an HTTP handler that upgrades connections to WebSocket.
func Handler(hub *Hub, jwtSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
			return
		}

		// Validate JWT and extract user ID
		userID, err := validateToken(jwtSecret, token)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade failed: %v", err)
			return
		}

		client := NewClient(userID, hub, conn)
		hub.Register <- client

		// Send authentication confirmation
		authMsg, _ := json.Marshal(map[string]string{
			"type":    "authenticated",
			"user_id": userID,
		})
		client.Send <- authMsg

		go client.WritePump()
		go client.ReadPump(newMessageHandler(hub))
	})
}

func validateToken(secret, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", jwt.ErrSignatureInvalid
	}

	return userID, nil
}

func newMessageHandler(hub *Hub) MessageHandler {
	return func(client *Client, msg IncomingMessage) {
		switch msg.Type {
		case "ping":
			client.Send <- []byte(`{"type":"pong"}`)

		case MsgTypeJoinRoom, MsgTypeSDPAnswer, MsgTypeICECandidate,
			MsgTypePublish, MsgTypeUnpublish, MsgTypeLeaveRoom:
			PublishSignal(hub.rmqPublisher, client.UserID, msg.RoomID, msg.Type, msg.Data)

		default:
			client.sendError("unknown_type", "unknown message type: "+msg.Type)
		}
	}
}
