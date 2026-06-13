package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/awwal/voxmeet/chat-service/internal/handler"
	"github.com/awwal/voxmeet/chat-service/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	createMsgFn func(ctx context.Context, arg store.CreateMessageParams) (store.Message, error)
	getMsgsFn   func(ctx context.Context, arg store.GetMessagesByRoomParams) ([]store.GetMessagesByRoomRow, error)
	deleteFn    func(ctx context.Context, arg store.DeleteMessageParams) error
}

func (m *mockStore) CreateMessage(ctx context.Context, arg store.CreateMessageParams) (store.Message, error) {
	return m.createMsgFn(ctx, arg)
}
func (m *mockStore) GetMessagesByRoom(ctx context.Context, arg store.GetMessagesByRoomParams) ([]store.GetMessagesByRoomRow, error) {
	return m.getMsgsFn(ctx, arg)
}
func (m *mockStore) DeleteMessage(ctx context.Context, arg store.DeleteMessageParams) error {
	return m.deleteFn(ctx, arg)
}

type mockPublisher struct {
	publishFn func(exchange, routingKey string, body []byte) error
}

func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return m.publishFn(exchange, routingKey, body)
}

func Test_HandleMessage_CreatesAndBroadcasts(t *testing.T) {
	var published []string
	pub := &mockPublisher{
		publishFn: func(exchange, routingKey string, body []byte) error {
			published = append(published, string(body))
			return nil
		},
	}
	st := &mockStore{
		createMsgFn: func(ctx context.Context, arg store.CreateMessageParams) (store.Message, error) {
			return store.Message{
				ID:        pgtype.UUID{Bytes: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Valid: true},
				RoomID:    arg.RoomID,
				UserID:    arg.UserID,
				Content:   arg.Content,
				Type:      arg.Type,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	h := handler.NewHandlerWithStore(st, pub)
	err := h.HandleMessage(context.Background(), handler.MessageEvent{
		RoomID:  "room-1",
		UserID:  "user-1",
		Content: "Hello!",
	})
	require.NoError(t, err)
	assert.Len(t, published, 1)
}

func Test_HandleMessage_EmptyContent(t *testing.T) {
	h := handler.NewHandlerWithStore(&mockStore{}, &mockPublisher{})
	err := h.HandleMessage(context.Background(), handler.MessageEvent{
		RoomID:  "room-1",
		UserID:  "user-1",
		Content: "",
	})
	assert.Error(t, err)
}

func Test_GetHistory_ReturnsMessages(t *testing.T) {
	st := &mockStore{
		getMsgsFn: func(ctx context.Context, arg store.GetMessagesByRoomParams) ([]store.GetMessagesByRoomRow, error) {
			return []store.GetMessagesByRoomRow{
				{Content: "msg1", CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
				{Content: "msg2", CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
			}, nil
		},
	}

	h := handler.NewHandlerWithStore(st, &mockPublisher{})
	msgs, err := h.GetHistory(context.Background(), "room-1", 50, 0)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
	assert.Equal(t, "msg1", msgs[0].Content)
}

func Test_GetHistory_StoreError(t *testing.T) {
	st := &mockStore{
		getMsgsFn: func(ctx context.Context, arg store.GetMessagesByRoomParams) ([]store.GetMessagesByRoomRow, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewHandlerWithStore(st, &mockPublisher{})
	_, err := h.GetHistory(context.Background(), "room-1", 50, 0)
	assert.Error(t, err)
}

func Test_DeleteMessage(t *testing.T) {
	var deleted bool
	st := &mockStore{
		deleteFn: func(ctx context.Context, arg store.DeleteMessageParams) error {
			deleted = true
			return nil
		},
	}

	h := handler.NewHandlerWithStore(st, &mockPublisher{})
	err := h.DeleteMessage(context.Background(), "room-1", "msg-1")
	require.NoError(t, err)
	assert.True(t, deleted)
}
