package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	prefix = "voxmeet"
)

type Config struct {
	URL                 string
	ReconnectInterval   time.Duration
	MaxReconnectAttempts int
}

func DefaultConfig() Config {
	return Config{
		URL:                  "amqp://guest:guest@localhost:5672/",
		ReconnectInterval:    30 * time.Second,
		MaxReconnectAttempts: 10,
	}
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return errors.New("rabbitmq URL is required")
	}
	if !strings.HasPrefix(c.URL, "amqp://") && !strings.HasPrefix(c.URL, "amqps://") {
		return fmt.Errorf("invalid rabbitmq URL: %s", c.URL)
	}
	return nil
}

type Connection struct {
	cfg  Config
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex
}

func New(cfg Config) *Connection {
	return &Connection{cfg: cfg}
}

func (c *Connection) Connect(ctx context.Context) error {
	if err := c.cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	conn, err := amqp.DialConfig(c.cfg.URL, amqp.Config{
		Heartbeat: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.ch = ch
	c.mu.Unlock()

	return nil
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Connection) DeclareExchange(name, kind string) error {
	c.mu.Lock()
	ch := c.ch
	c.mu.Unlock()

	if ch == nil {
		return errors.New("not connected")
	}

	return ch.ExchangeDeclare(
		name,
		kind,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
}

func (c *Connection) DeclareQueue(name string) (string, error) {
	c.mu.Lock()
	ch := c.ch
	c.mu.Unlock()

	if ch == nil {
		return "", errors.New("not connected")
	}

	if name == "" {
		q, err := ch.QueueDeclare(
			"",    // name (server-generated)
			false, // durable
			true,  // auto-delete
			true,  // exclusive
			false, // no-wait
			nil,
		)
		return q.Name, err
	}

	q, err := ch.QueueDeclare(
		name,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	return q.Name, err
}

func (c *Connection) BindQueue(queue, exchange, routingKey string) error {
	c.mu.Lock()
	ch := c.ch
	c.mu.Unlock()

	if ch == nil {
		return errors.New("not connected")
	}

	return ch.QueueBind(queue, routingKey, exchange, false, nil)
}

func (c *Connection) Consume(ctx context.Context, queue string) (<-chan amqp.Delivery, error) {
	c.mu.Lock()
	ch := c.ch
	c.mu.Unlock()

	if ch == nil {
		return nil, errors.New("not connected")
	}

	return ch.ConsumeWithContext(ctx, queue, "", false, false, false, false, nil)
}

func (c *Connection) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	c.mu.Lock()
	ch := c.ch
	c.mu.Unlock()

	if ch == nil {
		return errors.New("not connected")
	}

	return ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}

// ExchangeName returns the fully-qualified exchange name for the given key.
func ExchangeName(key string) string {
	return prefix + "." + key
}

// RoutingKey builds a routing key: "{exchange}.room.{roomID}".
func RoutingKey(exchange, roomID string) string {
	return exchange + ".room." + roomID
}
