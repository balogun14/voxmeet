package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/awwal/voxmeet/room-service/internal/handler"
	"github.com/awwal/voxmeet/room-service/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	createFn   func(ctx context.Context, arg store.CreateRoomParams) (store.Room, error)
	getFn      func(ctx context.Context, id pgtype.UUID) (store.Room, error)
	listFn     func(ctx context.Context) ([]store.Room, error)
	updateFn   func(ctx context.Context, arg store.UpdateRoomParams) (store.Room, error)
	deleteFn   func(ctx context.Context, id pgtype.UUID) error
	addMemberFn  func(ctx context.Context, arg store.AddRoomMemberParams) error
	removeMemberFn func(ctx context.Context, arg store.RemoveRoomMemberParams) error
	getMembersFn  func(ctx context.Context, roomID pgtype.UUID) ([]store.GetRoomMembersRow, error)
}

func (m *mockStore) CreateRoom(ctx context.Context, arg store.CreateRoomParams) (store.Room, error) { return m.createFn(ctx, arg) }
func (m *mockStore) GetRoomById(ctx context.Context, id pgtype.UUID) (store.Room, error) { return m.getFn(ctx, id) }
func (m *mockStore) ListRooms(ctx context.Context) ([]store.Room, error) { return m.listFn(ctx) }
func (m *mockStore) UpdateRoom(ctx context.Context, arg store.UpdateRoomParams) (store.Room, error) { return m.updateFn(ctx, arg) }
func (m *mockStore) DeleteRoom(ctx context.Context, id pgtype.UUID) error { return m.deleteFn(ctx, id) }
func (m *mockStore) AddRoomMember(ctx context.Context, arg store.AddRoomMemberParams) error { return m.addMemberFn(ctx, arg) }
func (m *mockStore) RemoveRoomMember(ctx context.Context, arg store.RemoveRoomMemberParams) error { return m.removeMemberFn(ctx, arg) }
func (m *mockStore) GetRoomMembers(ctx context.Context, roomID pgtype.UUID) ([]store.GetRoomMembersRow, error) { return m.getMembersFn(ctx, roomID) }
func (m *mockStore) ListRoomsByUser(ctx context.Context, userID pgtype.UUID) ([]store.Room, error) { return nil, nil }
func (m *mockStore) IsRoomMember(ctx context.Context, arg store.IsRoomMemberParams) (bool, error) { return false, nil }

func Test_Handler_CreateRoom(t *testing.T) {
	h := handler.NewHandler(&mockStore{
		createFn: func(ctx context.Context, arg store.CreateRoomParams) (store.Room, error) {
			return store.Room{ID: pgtype.UUID{Bytes: uuid.New(), Valid: true}, Name: arg.Name, OwnerID: arg.OwnerID, CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}, nil
		},
	})
	r, err := h.CreateRoom(context.Background(), "user-1", "My Room", true, 10)
	require.NoError(t, err)
	assert.Equal(t, "My Room", r.Name)
}

func Test_Handler_GetRoom(t *testing.T) {
	roomID := uuid.New()
	h := handler.NewHandler(&mockStore{
		getFn: func(ctx context.Context, id pgtype.UUID) (store.Room, error) {
			return store.Room{ID: id, Name: "Found Room"}, nil
		},
	})
	r, err := h.GetRoom(context.Background(), roomID.String())
	require.NoError(t, err)
	assert.Equal(t, "Found Room", r.Name)
}

func Test_Handler_GetRoom_NotFound(t *testing.T) {
	h := handler.NewHandler(&mockStore{
		getFn: func(ctx context.Context, id pgtype.UUID) (store.Room, error) {
			return store.Room{}, errors.New("not found")
		},
	})
	_, err := h.GetRoom(context.Background(), uuid.New().String())
	assert.Error(t, err)
}

func Test_Handler_ListRooms(t *testing.T) {
	h := handler.NewHandler(&mockStore{
		listFn: func(ctx context.Context) ([]store.Room, error) {
			return []store.Room{{Name: "Room A"}, {Name: "Room B"}}, nil
		},
	})
	rooms, err := h.ListRooms(context.Background())
	require.NoError(t, err)
	assert.Len(t, rooms, 2)
}

func Test_Handler_DeleteRoom(t *testing.T) {
	var deleted bool
	h := handler.NewHandler(&mockStore{
		deleteFn: func(ctx context.Context, id pgtype.UUID) error { deleted = true; return nil },
	})
	err := h.DeleteRoom(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.True(t, deleted)
}

func Test_Handler_AddMember(t *testing.T) {
	var added bool
	h := handler.NewHandler(&mockStore{
		addMemberFn: func(ctx context.Context, arg store.AddRoomMemberParams) error { added = true; return nil },
	})
	err := h.AddMember(context.Background(), uuid.New().String(), uuid.New().String(), "member")
	require.NoError(t, err)
	assert.True(t, added)
}

func Test_Handler_RemoveMember(t *testing.T) {
	var removed bool
	h := handler.NewHandler(&mockStore{
		removeMemberFn: func(ctx context.Context, arg store.RemoveRoomMemberParams) error { removed = true; return nil },
	})
	err := h.RemoveMember(context.Background(), uuid.New().String(), uuid.New().String())
	require.NoError(t, err)
	assert.True(t, removed)
}

func Test_Handler_GetMembers(t *testing.T) {
	h := handler.NewHandler(&mockStore{
		getMembersFn: func(ctx context.Context, roomID pgtype.UUID) ([]store.GetRoomMembersRow, error) {
			return []store.GetRoomMembersRow{{Username: "alice"}, {Username: "bob"}}, nil
		},
	})
	members, err := h.GetMembers(context.Background(), uuid.New().String())
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "alice", members[0].Username)
}
