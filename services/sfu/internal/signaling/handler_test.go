package signaling_test

import (
	"context"
	"testing"

	"github.com/awwal/voxmeet/sfu/internal/signaling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Producer_NilConnection(t *testing.T) {
	p := signaling.NewProducer(nil)
	assert.NotNil(t, p)
}

func Test_Producer_Publish(t *testing.T) {
	p := signaling.NewProducer(&mockRMQ{})
	err := p.Publish(context.Background(), signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-1",
	})
	assert.NoError(t, err)
}

func Test_NewConsumer_NilConnection(t *testing.T) {
	_, err := signaling.NewConsumer(nil, signaling.NewProducer(nil))
	assert.Error(t, err)
}

func Test_Consumer_Start_CancelledContext(t *testing.T) {
	consumer, err := signaling.NewConsumer(&mockRMQ{}, signaling.NewProducer(&mockRMQ{}))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return without hanging
	err = consumer.Start(ctx)
	// May return nil or an error depending on when context is checked
	t.Logf("Start returned: %v", err)
}

// mockRMQ implements the RMQConn interface for testing.
type mockRMQ struct{}

func (m *mockRMQ) Connect(ctx context.Context) error { return nil }
func (m *mockRMQ) Close() error { return nil }
func (m *mockRMQ) DeclareExchange(name, kind string) error { return nil }
func (m *mockRMQ) DeclareQueue(name string) (string, error) { return "test-queue", nil }
func (m *mockRMQ) BindQueue(queue, exchange, routingKey string) error { return nil }
func (m *mockRMQ) Consume(ctx context.Context, queue string) (<-chan []byte, error) {
	ch := make(chan []byte)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}
func (m *mockRMQ) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return nil
}
