package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/awwal/voxmeet/api-gateway/internal/auth"
	"github.com/awwal/voxmeet/api-gateway/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type RoomHandler struct {
	queries db.Querier
}

func NewRoomHandler(queries db.Querier) *RoomHandler {
	return &RoomHandler{queries: queries}
}

type createRoomRequest struct {
	Name            string `json:"name"`
	IsPublic        bool   `json:"is_public"`
	MaxParticipants int    `json:"max_participants"`
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "room name is required")
		return
	}
	if req.MaxParticipants <= 0 {
		req.MaxParticipants = 20
	}

	uid, _ := uuid.Parse(userID)

	room, err := h.queries.CreateRoom(r.Context(), db.CreateRoomParams{
		Name:            req.Name,
		OwnerID:         pgtype.UUID{Bytes: uid, Valid: true},
		IsPublic:        req.IsPublic,
		MaxParticipants: int32(req.MaxParticipants),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create room")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":               userIDToString(room.ID),
		"name":             room.Name,
		"owner_id":         userIDToString(room.OwnerID),
		"is_public":        room.IsPublic,
		"max_participants": room.MaxParticipants,
		"created_at":       room.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.queries.ListRooms(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list rooms")
		return
	}

	result := make([]map[string]interface{}, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, map[string]interface{}{
			"id":               userIDToString(room.ID),
			"name":             room.Name,
			"owner_id":         userIDToString(room.OwnerID),
			"is_public":        room.IsPublic,
			"max_participants": room.MaxParticipants,
			"created_at":       room.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	respondJSON(w, http.StatusOK, result)
}

func (h *RoomHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	uid, err := uuid.Parse(roomID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	room, err := h.queries.GetRoomById(r.Context(), pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		respondError(w, http.StatusNotFound, "room not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":               userIDToString(room.ID),
		"name":             room.Name,
		"owner_id":         userIDToString(room.OwnerID),
		"is_public":        room.IsPublic,
		"max_participants": room.MaxParticipants,
		"created_at":       room.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

type updateRoomRequest struct {
	Name            string `json:"name"`
	IsPublic        bool   `json:"is_public"`
	MaxParticipants int    `json:"max_participants"`
}

func (h *RoomHandler) Update(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	uid, err := uuid.Parse(roomID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	// Fetch existing to get fallback values
	existing, err := h.queries.GetRoomById(r.Context(), pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		respondError(w, http.StatusNotFound, "room not found")
		return
	}

	var req updateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = existing.Name
	}
	if req.MaxParticipants <= 0 {
		req.MaxParticipants = int(existing.MaxParticipants)
	}

	room, err := h.queries.UpdateRoom(r.Context(), db.UpdateRoomParams{
		ID:              pgtype.UUID{Bytes: uid, Valid: true},
		Name:            req.Name,
		IsPublic:        req.IsPublic,
		MaxParticipants: int32(req.MaxParticipants),
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "room not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":               userIDToString(room.ID),
		"name":             room.Name,
		"owner_id":         userIDToString(room.OwnerID),
		"is_public":        room.IsPublic,
		"max_participants": room.MaxParticipants,
		"created_at":       room.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	uid, err := uuid.Parse(roomID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	if err := h.queries.DeleteRoom(r.Context(), pgtype.UUID{Bytes: uid, Valid: true}); err != nil {
		respondError(w, http.StatusNotFound, "room not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
