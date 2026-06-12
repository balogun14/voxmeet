package rabbitmq_test

import (
	"context"
	"testing"
	"time"

	"github.com/awwal/voxmeet/pkgs/rabbitmq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Config_Defaults(t *testing.T) {
	cfg := rabbitmq.DefaultConfig()
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.URL)
	assert.Equal(t, 30*time.Second, cfg.ReconnectInterval)
	assert.Equal(t, 10, cfg.MaxReconnectAttempts)
}

func Test_Config_InvalidURL(t *testing.T) {
	cfg := rabbitmq.Config{URL: "not-a-valid-url"}
	err := cfg.Validate()
	assert.Error(t, err)
}

func Test_Config_ValidURL(t *testing.T) {
	cfg := rabbitmq.Config{URL: "amqp://user:pass@host:5672/vhost"}
	err := cfg.Validate()
	assert.NoError(t, err)
}

func Test_Exchange_Names(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"signal", "voxmeet.signal"},
		{"chat", "voxmeet.chat"},
		{"room", "voxmeet.room"},
		{"presence", "voxmeet.presence"},
		{"rpc", "voxmeet.rpc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rabbitmq.ExchangeName(tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_RoutingKey(t *testing.T) {
	tests := []struct {
		exchange string
		roomID   string
		want     string
	}{
		{"signal", "room-abc", "signal.room.room-abc"},
		{"chat", "room-abc", "chat.room.room-abc"},
		{"presence", "room-abc", "presence.room.room-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.exchange+"_"+tt.roomID, func(t *testing.T) {
			got := rabbitmq.RoutingKey(tt.exchange, tt.roomID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_NewConnection_InvalidURL(t *testing.T) {
	cfg := rabbitmq.Config{URL: "amqp://invalid"}
	conn := rabbitmq.New(cfg)
	assert.NotNil(t, conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := conn.Connect(ctx)
	assert.Error(t, err)
}

func Test_NewConnection_NilConfig(t *testing.T) {
	conn := rabbitmq.New(rabbitmq.Config{})
	assert.NotNil(t, conn)
}

// Integration test — requires a running RabbitMQ instance.
// Run with: go test -tags=integration ./pkgs/rabbitmq/
func Test_Integration_ConnectAndPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg := rabbitmq.DefaultConfig()
	conn := rabbitmq.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := conn.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	err = conn.DeclareExchange("voxmeet.test", "topic")
	require.NoError(t, err)

	queue, err := conn.DeclareQueue("test-queue")
	require.NoError(t, err)

	err = conn.BindQueue(queue, "voxmeet.test", "test.#")
	require.NoError(t, err)

	msgs, err := conn.Consume(ctx, queue)
	require.NoError(t, err)

	err = conn.Publish(ctx, "voxmeet.test", "test.message", []byte("hello"))
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "hello", string(msg.Body))
	case <-ctx.Done():
		t.Fatal("timeout waiting for message")
	}
}
