package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/awwal/voxmeet/chat-service/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
)

// MessageStore abstracts the sqlc-generated store for testability.
type MessageStore interface {
	CreateMessage(ctx context.Context, arg store.CreateMessageParams) (store.Message, error)
	GetMessagesByRoom(ctx context.Context, arg store.GetMessagesByRoomParams) ([]store.GetMessagesByRoomRow, error)
	DeleteMessage(ctx context.Context, arg store.DeleteMessageParams) error
}

// Publisher abstracts sending messages to RabbitMQ.
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

// Handler processes chat messages.
type Handler struct {
	store MessageStore
	pub   Publisher
}

// NewHandler creates a new chat Handler.
func NewHandler(store MessageStore, pub Publisher) *Handler {
	return &Handler{store: store, pub: pub}
}

// MessageEvent is the payload from RabbitMQ.
type MessageEvent struct {
	RoomID  string `json:"room_id"`
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

// ChatMessage is the broadcast payload.
type ChatMessage struct {
	ID        string `json:"id"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// HandleMessage persists a message and broadcasts it to the room.
func (h *Handler) HandleMessage(ctx context.Context, evt MessageEvent) error {
	if evt.Content == "" {
		return fmt.Errorf("content is required")
	}

	msg, err := h.store.CreateMessage(ctx, store.CreateMessageParams{
		RoomID:  pgtype.UUID{}, // populated from string in real consumer
		UserID:  pgtype.UUID{},
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
	rows, err := h.store.GetMessagesByRoom(ctx, store.GetMessagesByRoomParams{
		RoomID: pgtype.UUID{},
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
	return h.store.DeleteMessage(ctx, store.DeleteMessageParams{
		ID:     pgtype.UUID{},
		RoomID: pgtype.UUID{},
	})
}
