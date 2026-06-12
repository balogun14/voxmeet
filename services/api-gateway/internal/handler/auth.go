package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/awwal/voxmeet/api-gateway/internal/auth"
	"github.com/awwal/voxmeet/api-gateway/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthHandler struct {
	queries   db.Querier
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthHandler(queries db.Querier, jwtSecret string, jwtTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		queries:   queries,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token    string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}
	if len(req.Password) < 6 {
		respondError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}
	if !strings.Contains(req.Email, "@") {
		respondError(w, http.StatusBadRequest, "invalid email address")
		return
	}

	if _, err := h.queries.GetUserByEmail(r.Context(), req.Email); err == nil {
		respondError(w, http.StatusConflict, "email already registered")
		return
	}

	if _, err := h.queries.GetUserByUsername(r.Context(), req.Username); err == nil {
		respondError(w, http.StatusConflict, "username already taken")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Username:    req.Username,
		Email:       req.Email,
		Password:    hash,
		DisplayName: pgtype.Text{String: req.Username, Valid: true},
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, expiresAt, err := auth.GenerateJWT(h.jwtSecret, userIDToString(user.ID), h.jwtTTL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respondJSON(w, http.StatusCreated, authResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		UserID:     userIDToString(user.ID),
		Username:   user.Username,
		Email:      user.Email,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !auth.CheckPassword(user.Password, req.Password) {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, expiresAt, err := auth.GenerateJWT(h.jwtSecret, userIDToString(user.ID), h.jwtTTL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, authResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		UserID:     userIDToString(user.ID),
		Username:   user.Username,
		Email:      user.Email,
	})
}

func userIDToString(id pgtype.UUID) string {
	uid, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return ""
	}
	return uid.String()
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// ErrNotFound is a sentinel error for "not found" cases.
var ErrNotFound = errors.New("not found")
