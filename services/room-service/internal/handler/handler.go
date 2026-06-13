package handler

import (
	"context"
	"fmt"

	"github.com/awwal/voxmeet/room-service/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// RoomStore abstracts the sqlc-generated store for testability.
type RoomStore interface {
	CreateRoom(ctx context.Context, arg store.CreateRoomParams) (store.Room, error)
	GetRoomById(ctx context.Context, id pgtype.UUID) (store.Room, error)
	ListRooms(ctx context.Context) ([]store.Room, error)
	UpdateRoom(ctx context.Context, arg store.UpdateRoomParams) (store.Room, error)
	DeleteRoom(ctx context.Context, id pgtype.UUID) error
	AddRoomMember(ctx context.Context, arg store.AddRoomMemberParams) error
	RemoveRoomMember(ctx context.Context, arg store.RemoveRoomMemberParams) error
	GetRoomMembers(ctx context.Context, roomID pgtype.UUID) ([]store.GetRoomMembersRow, error)
}

// Handler processes room operations.
type Handler struct {
	store RoomStore
}

// NewHandler creates a new room Handler.
func NewHandler(q *store.Queries) *Handler {
	return &Handler{store: &queriesStore{q: q}}
}

// NewHandlerWithStore creates a room Handler with a custom store (for testing).
func NewHandlerWithStore(s RoomStore) *Handler {
	return &Handler{store: s}
}

// queriesStore adapts store.Queries to RoomStore.
type queriesStore struct {
	q *store.Queries
}

func (s *queriesStore) CreateRoom(ctx context.Context, arg store.CreateRoomParams) (store.Room, error) { return s.q.CreateRoom(ctx, arg) }
func (s *queriesStore) GetRoomById(ctx context.Context, id pgtype.UUID) (store.Room, error) { return s.q.GetRoomById(ctx, id) }
func (s *queriesStore) ListRooms(ctx context.Context) ([]store.Room, error) { return s.q.ListRooms(ctx) }
func (s *queriesStore) UpdateRoom(ctx context.Context, arg store.UpdateRoomParams) (store.Room, error) { return s.q.UpdateRoom(ctx, arg) }
func (s *queriesStore) DeleteRoom(ctx context.Context, id pgtype.UUID) error { return s.q.DeleteRoom(ctx, id) }
func (s *queriesStore) AddRoomMember(ctx context.Context, arg store.AddRoomMemberParams) error { return s.q.AddRoomMember(ctx, arg) }
func (s *queriesStore) RemoveRoomMember(ctx context.Context, arg store.RemoveRoomMemberParams) error { return s.q.RemoveRoomMember(ctx, arg) }
func (s *queriesStore) GetRoomMembers(ctx context.Context, roomID pgtype.UUID) ([]store.GetRoomMembersRow, error) { return s.q.GetRoomMembers(ctx, roomID) }

func (h *Handler) CreateRoom(ctx context.Context, ownerID, name string, isPublic bool, maxParticipants int32) (*store.Room, error) {
	uid, _ := uuid.Parse(ownerID)
	room, err := h.store.CreateRoom(ctx, store.CreateRoomParams{
		Name:            name,
		OwnerID:         pgtype.UUID{Bytes: uid, Valid: true},
		IsPublic:        isPublic,
		MaxParticipants: maxParticipants,
	})
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return &room, nil
}

func (h *Handler) GetRoom(ctx context.Context, roomID string) (*store.Room, error) {
	uid, err := uuid.Parse(roomID)
	if err != nil {
		return nil, fmt.Errorf("invalid room id: %w", err)
	}
	room, err := h.store.GetRoomById(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}
	return &room, nil
}

func (h *Handler) ListRooms(ctx context.Context) ([]store.Room, error) {
	return h.store.ListRooms(ctx)
}

func (h *Handler) DeleteRoom(ctx context.Context, roomID string) error {
	uid, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("invalid room id: %w", err)
	}
	return h.store.DeleteRoom(ctx, pgtype.UUID{Bytes: uid, Valid: true})
}

func (h *Handler) AddMember(ctx context.Context, roomID, userID, role string) error {
	rID, _ := uuid.Parse(roomID)
	uID, _ := uuid.Parse(userID)
	return h.store.AddRoomMember(ctx, store.AddRoomMemberParams{
		RoomID: pgtype.UUID{Bytes: rID, Valid: true},
		UserID: pgtype.UUID{Bytes: uID, Valid: true},
		Role:   role,
	})
}

func (h *Handler) RemoveMember(ctx context.Context, roomID, userID string) error {
	rID, _ := uuid.Parse(roomID)
	uID, _ := uuid.Parse(userID)
	return h.store.RemoveRoomMember(ctx, store.RemoveRoomMemberParams{
		RoomID: pgtype.UUID{Bytes: rID, Valid: true},
		UserID: pgtype.UUID{Bytes: uID, Valid: true},
	})
}

func (h *Handler) GetMembers(ctx context.Context, roomID string) ([]store.GetRoomMembersRow, error) {
	rID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, fmt.Errorf("invalid room id: %w", err)
	}
	return h.store.GetRoomMembers(ctx, pgtype.UUID{Bytes: rID, Valid: true})
}
