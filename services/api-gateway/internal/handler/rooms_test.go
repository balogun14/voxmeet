package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/awwal/voxmeet/api-gateway/internal/auth"
	"github.com/awwal/voxmeet/api-gateway/internal/handler"
	"github.com/awwal/voxmeet/api-gateway/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func roomUUID(s string) pgtype.UUID {
	id, _ := uuid.Parse(s)
	return pgtype.UUID{Bytes: id, Valid: true}
}

func Test_RoomHandler_Create(t *testing.T) {
	mock := &mockQuerier{
		createRoomFn: func(ctx context.Context, arg db.CreateRoomParams) (db.Room, error) {
			return db.Room{
				ID:              roomUUID(uuid.New().String()),
				Name:            arg.Name,
				OwnerID:         arg.OwnerID,
				IsPublic:        arg.IsPublic,
				MaxParticipants: arg.MaxParticipants,
				CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	h := handler.NewRoomHandler(mock)
	userID := uuid.New().String()
	token, _, err := auth.GenerateJWT("test-secret", userID, 1*time.Hour)
	require.NoError(t, err)

	body := `{"name":"Team Standup","is_public":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Inject user context
	ctx := auth.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Team Standup", resp["name"])
	assert.NotEmpty(t, resp["id"])
}

func Test_RoomHandler_Create_MissingName(t *testing.T) {
	mock := &mockQuerier{}
	h := handler.NewRoomHandler(mock)
	userID := uuid.New().String()
	ctx := auth.WithUserID(context.Background(), userID)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.Create(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_RoomHandler_List(t *testing.T) {
	now := time.Now()
	rooms := []db.Room{
		{
			ID:              roomUUID(uuid.New().String()),
			Name:            "Room 1",
			OwnerID:         roomUUID(uuid.New().String()),
			IsPublic:        true,
			MaxParticipants: 10,
			CreatedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			ID:              roomUUID(uuid.New().String()),
			Name:            "Room 2",
			OwnerID:         roomUUID(uuid.New().String()),
			IsPublic:        false,
			MaxParticipants: 20,
			CreatedAt:       pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true},
		},
	}

	mock := &mockQuerier{
		listRoomsFn: func(ctx context.Context) ([]db.Room, error) { return rooms, nil },
	}

	h := handler.NewRoomHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "Room 1", resp[0]["name"])
}

func Test_RoomHandler_GetByID(t *testing.T) {
	roomID := uuid.New().String()
	ownerID := uuid.New().String()

	mock := &mockQuerier{
		getRoomByIdFn: func(ctx context.Context, id pgtype.UUID) (db.Room, error) {
			return db.Room{
				ID:       roomUUID(roomID),
				Name:     "Specific Room",
				OwnerID:  roomUUID(ownerID),
				IsPublic: true,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	h := handler.NewRoomHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/"+roomID, nil)
	req.SetPathValue("id", roomID)
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Specific Room", resp["name"])
	assert.Equal(t, roomID, resp["id"])
}

func Test_RoomHandler_GetByID_NotFound(t *testing.T) {
	roomID := uuid.New().String()
	mock := &mockQuerier{
		getRoomByIdFn: func(ctx context.Context, id pgtype.UUID) (db.Room, error) {
			return db.Room{}, errors.New("not found")
		},
	}

	h := handler.NewRoomHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/"+roomID, nil)
	req.SetPathValue("id", roomID)
	w := httptest.NewRecorder()

	h.GetByID(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
