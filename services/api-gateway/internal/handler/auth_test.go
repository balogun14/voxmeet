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

// mockQuerier implements db.Querier for testing.
type mockQuerier struct {
	createUserFn        func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	getUserByEmailFn    func(ctx context.Context, email string) (db.User, error)
	getUserByUsernameFn func(ctx context.Context, username string) (db.User, error)
	createSessionFn     func(ctx context.Context, arg db.CreateSessionParams) (db.Session, error)
	createRoomFn        func(ctx context.Context, arg db.CreateRoomParams) (db.Room, error)
	listRoomsFn         func(ctx context.Context) ([]db.Room, error)
	getRoomByIdFn       func(ctx context.Context, id pgtype.UUID) (db.Room, error)
}

func (m *mockQuerier) AddRoomMember(ctx context.Context, arg db.AddRoomMemberParams) error { return nil }
func (m *mockQuerier) CreateMessage(ctx context.Context, arg db.CreateMessageParams) (db.Message, error) { return db.Message{}, nil }
func (m *mockQuerier) CreateRoom(ctx context.Context, arg db.CreateRoomParams) (db.Room, error) {
	if m.createRoomFn != nil {
		return m.createRoomFn(ctx, arg)
	}
	return db.Room{}, nil
}
func (m *mockQuerier) DeleteMessage(ctx context.Context, arg db.DeleteMessageParams) error { return nil }
func (m *mockQuerier) DeleteRoom(ctx context.Context, id pgtype.UUID) error { return nil }
func (m *mockQuerier) DeleteSession(ctx context.Context, token string) error { return nil }
func (m *mockQuerier) DeleteUserSessions(ctx context.Context, userID pgtype.UUID) error { return nil }
func (m *mockQuerier) GetMessagesByRoom(ctx context.Context, arg db.GetMessagesByRoomParams) ([]db.GetMessagesByRoomRow, error) { return nil, nil }
func (m *mockQuerier) GetRoomById(ctx context.Context, id pgtype.UUID) (db.Room, error) {
	if m.getRoomByIdFn != nil {
		return m.getRoomByIdFn(ctx, id)
	}
	return db.Room{}, errors.New("not found")
}
func (m *mockQuerier) GetRoomMembers(ctx context.Context, roomID pgtype.UUID) ([]db.GetRoomMembersRow, error) { return nil, nil }
func (m *mockQuerier) GetSessionByToken(ctx context.Context, token string) (db.GetSessionByTokenRow, error) { return db.GetSessionByTokenRow{}, nil }
func (m *mockQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) { return m.getUserByEmailFn(ctx, email) }
func (m *mockQuerier) GetUserByUsername(ctx context.Context, username string) (db.User, error) { return m.getUserByUsernameFn(ctx, username) }
func (m *mockQuerier) IsRoomMember(ctx context.Context, arg db.IsRoomMemberParams) (bool, error) { return false, nil }
func (m *mockQuerier) ListRooms(ctx context.Context) ([]db.Room, error) {
	if m.listRoomsFn != nil {
		return m.listRoomsFn(ctx)
	}
	return nil, nil
}
func (m *mockQuerier) ListRoomsByUser(ctx context.Context, userID pgtype.UUID) ([]db.Room, error) { return nil, nil }
func (m *mockQuerier) RemoveRoomMember(ctx context.Context, arg db.RemoveRoomMemberParams) error { return nil }
func (m *mockQuerier) UpdateRoom(ctx context.Context, arg db.UpdateRoomParams) (db.Room, error) { return db.Room{}, nil }
func (m *mockQuerier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) { return db.User{}, nil }
func (m *mockQuerier) GetUserById(ctx context.Context, id pgtype.UUID) (db.User, error) { return db.User{}, nil }
func (m *mockQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) { return m.createUserFn(ctx, arg) }
func (m *mockQuerier) CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) { return m.createSessionFn(ctx, arg) }

func uuidToPG(s string) pgtype.UUID {
	id, _ := uuid.Parse(s)
	return pgtype.UUID{Bytes: id, Valid: true}
}

func Test_AuthHandler_Register_Success(t *testing.T) {
	mock := &mockQuerier{
		getUserByEmailFn: func(ctx context.Context, email string) (db.User, error) {
			return db.User{}, errors.New("not found")
		},
		getUserByUsernameFn: func(ctx context.Context, username string) (db.User, error) {
			return db.User{}, errors.New("not found")
		},
		createUserFn: func(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{
				ID:          uuidToPG(uuid.New().String()),
				Username:    arg.Username,
				Email:       arg.Email,
				DisplayName: pgtype.Text{String: arg.DisplayName.String, Valid: arg.DisplayName.Valid},
				CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
		createSessionFn: func(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
			return db.Session{
				Token: arg.Token,
				ExpiresAt: pgtype.Timestamptz{Time: arg.ExpiresAt.Time, Valid: true},
			}, nil
		},
	}

	h := handler.NewAuthHandler(mock, "test-secret", 24*time.Hour)

	body := `{"username":"alice","email":"alice@example.com","password":"secure123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "alice", resp["username"])
	assert.Equal(t, "alice@example.com", resp["email"])
	assert.NotEmpty(t, resp["token"])
}

func Test_AuthHandler_Register_DuplicateEmail(t *testing.T) {
	mock := &mockQuerier{
		getUserByEmailFn: func(ctx context.Context, email string) (db.User, error) {
			return db.User{
				ID:       uuidToPG(uuid.New().String()),
				Username: "alice",
				Email:    email,
			}, nil
		},
	}

	h := handler.NewAuthHandler(mock, "test-secret", 24*time.Hour)

	body := `{"username":"bob","email":"alice@example.com","password":"secure123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func Test_AuthHandler_Register_InvalidInput(t *testing.T) {
	mock := &mockQuerier{}
	h := handler.NewAuthHandler(mock, "test-secret", 24*time.Hour)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"missing username", `{"email":"a@b.com","password":"pass"}`},
		{"missing email", `{"username":"a","password":"pass"}`},
		{"missing password", `{"username":"a","email":"a@b.com"}`},
		{"short password", `{"username":"a","email":"a@b.com","password":"123"}`},
		{"invalid email", `{"username":"a","email":"not-email","password":"validPass1"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Register(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func Test_AuthHandler_Login_Success(t *testing.T) {
	hash, _ := auth.HashPassword("secure123")
	mock := &mockQuerier{
		getUserByEmailFn: func(ctx context.Context, email string) (db.User, error) {
			return db.User{
				ID:       uuidToPG(uuid.New().String()),
				Username: "alice",
				Email:    email,
				Password: hash,
			}, nil
		},
		createSessionFn: func(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
			return db.Session{
				Token: arg.Token,
				ExpiresAt: pgtype.Timestamptz{Time: arg.ExpiresAt.Time, Valid: true},
			}, nil
		},
	}

	h := handler.NewAuthHandler(mock, "test-secret", 24*time.Hour)

	body := `{"email":"alice@example.com","password":"secure123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "alice", resp["username"])
	assert.NotEmpty(t, resp["token"])
}

func Test_AuthHandler_Login_WrongPassword(t *testing.T) {
	hash, _ := auth.HashPassword("secure123")
	mock := &mockQuerier{
		getUserByEmailFn: func(ctx context.Context, email string) (db.User, error) {
			return db.User{
				ID:       uuidToPG(uuid.New().String()),
				Username: "alice",
				Email:    email,
				Password: hash,
			}, nil
		},
	}

	h := handler.NewAuthHandler(mock, "test-secret", 24*time.Hour)

	body := `{"email":"alice@example.com","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_AuthHandler_Login_UserNotFound(t *testing.T) {
	mock := &mockQuerier{
		getUserByEmailFn: func(ctx context.Context, email string) (db.User, error) {
			return db.User{}, errors.New("not found")
		},
	}

	h := handler.NewAuthHandler(mock, "test-secret", 24*time.Hour)

	body := `{"email":"nonexistent@example.com","password":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
