package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/awwal/voxmeet/chat-service/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Publisher abstracts sending messages to RabbitMQ.
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

// Store abstracts the sqlc-generated store for testability.
type Store interface {
	CreateMessage(ctx context.Context, arg store.CreateMessageParams) (store.Message, error)
	GetMessagesByRoom(ctx context.Context, arg store.GetMessagesByRoomParams) ([]store.GetMessagesByRoomRow, error)
	DeleteMessage(ctx context.Context, arg store.DeleteMessageParams) error
}

type queriesStore struct {
	q *store.Queries
}

func (s *queriesStore) CreateMessage(ctx context.Context, arg store.CreateMessageParams) (store.Message, error) {
	return s.q.CreateMessage(ctx, arg)
}
func (s *queriesStore) GetMessagesByRoom(ctx context.Context, arg store.GetMessagesByRoomParams) ([]store.GetMessagesByRoomRow, error) {
	return s.q.GetMessagesByRoom(ctx, arg)
}
func (s *queriesStore) DeleteMessage(ctx context.Context, arg store.DeleteMessageParams) error {
	return s.q.DeleteMessage(ctx, arg)
}

// Handler processes chat messages.
type Handler struct {
	store Store
	pub   Publisher
}

// NewHandler creates a new chat Handler.
func NewHandler(pool *pgxpool.Pool, pub Publisher) *Handler {
	return &Handler{
		store: &queriesStore{q: store.New(pool)},
		pub:   pub,
	}
}

// NewHandlerWithStore creates a chat Handler with a specific store (for testing).
func NewHandlerWithStore(s Store, pub Publisher) *Handler {
	return &Handler{store: s, pub: pub}
}

// MessageEvent is the payload from RabbitMQ.
type MessageEvent struct {
	RoomID  string `json:"room_id"`
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

// HandleMessage persists a message and broadcasts it to the room.
func (h *Handler) HandleMessage(ctx context.Context, evt MessageEvent) error {
	if evt.Content == "" {
		return fmt.Errorf("content is required")
	}

	roomUUID, _ := uuid.Parse(evt.RoomID)
	userUUID, _ := uuid.Parse(evt.UserID)

	msg, err := h.store.CreateMessage(ctx, store.CreateMessageParams{
		RoomID:  pgtype.UUID{Bytes: roomUUID, Valid: true},
		UserID:  pgtype.UUID{Bytes: userUUID, Valid: true},
		Content: evt.Content,
		Type:    "text",
	})
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"type":    "chat.message",
		"room_id": evt.RoomID,
		"user_id": evt.UserID,
		"content": msg.Content,
		"id":      msg.ID.String(),
		"created_at": msg.CreatedAt.Time.Format(time.RFC3339),
	})

	return h.pub.Publish(ctx, "voxmeet.chat", "chat.room."+evt.RoomID, body)
}

// GetHistory retrieves message history for a room.
func (h *Handler) GetHistory(ctx context.Context, roomID string, limit, offset int32) ([]ChatMessage, error) {
	roomUUID, _ := uuid.Parse(roomID)

	rows, err := h.store.GetMessagesByRoom(ctx, store.GetMessagesByRoomParams{
		RoomID: pgtype.UUID{Bytes: roomUUID, Valid: true},
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	msgs := make([]ChatMessage, 0, len(rows))
	for _, r := range rows {
		msgs = append(msgs, ChatMessage{
			ID:        r.ID.String(),
			RoomID:    r.RoomID.String(),
			UserID:    r.UserID.String(),
			Content:   r.Content,
			Type:      r.Type,
			CreatedAt: r.CreatedAt.Time.Format(time.RFC3339),
		})
	}
	return msgs, nil
}

// DeleteMessage deletes a message.
func (h *Handler) DeleteMessage(ctx context.Context, roomID, messageID string) error {
	roomUUID, _ := uuid.Parse(roomID)
	msgUUID, _ := uuid.Parse(messageID)
	return h.store.DeleteMessage(ctx, store.DeleteMessageParams{
		ID:     pgtype.UUID{Bytes: msgUUID, Valid: true},
		RoomID: pgtype.UUID{Bytes: roomUUID, Valid: true},
	})
}

// ChatMessage is the API response type.
type ChatMessage struct {
	ID        string `json:"id"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}
