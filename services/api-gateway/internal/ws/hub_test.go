package ws_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/awwal/voxmeet/api-gateway/internal/ws"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Hub_RegisterAndUnregister(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	assert.Equal(t, 0, hub.ClientCount())

	client := ws.NewClient("user-1", hub, nil)
	hub.Register <- client

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	hub.Unregister <- client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
}

func Test_Upgrade_NoToken(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	handler := ws.Handler(hub, "test-secret")

	server := httptest.NewServer(handler)
	defer server.Close()

	// Connect with no token — should get 401
	url := "ws://" + server.Listener.Addr().String() + "/api/v1/ws"
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func Test_Upgrade_WithValidToken(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	handler := ws.Handler(hub, "test-secret")

	server := httptest.NewServer(handler)
	defer server.Close()

	token, _, err := createTestJWT("test-secret", "user-1", time.Hour)
	require.NoError(t, err)

	url := "ws://" + server.Listener.Addr().String() + "/api/v1/ws?token=" + token
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)

	var evt map[string]interface{}
	err = json.Unmarshal(msg, &evt)
	require.NoError(t, err)
	assert.Equal(t, "authenticated", evt["type"])
	assert.Equal(t, "user-1", evt["user_id"])
}

func Test_Hub_ClientCount(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	assert.Equal(t, 0, hub.ClientCount())

	client1 := ws.NewClient("user-1", hub, nil)
	client2 := ws.NewClient("user-2", hub, nil)

	hub.Register <- client1
	hub.Register <- client2

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 2, hub.ClientCount())
}

func createTestJWT(secret, userID string, ttl time.Duration) (string, time.Time, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     "voxmeet",
		"sub":     userID,
		"jti":     "test-jti",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, time.Now().Add(ttl), nil
}
