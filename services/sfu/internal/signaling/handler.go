package signaling

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/awwal/voxmeet/pkgs/rabbitmq"
)

// Broker abstracts the RabbitMQ operations needed by the signaling system.
// This allows the signal package to be tested without a real RabbitMQ connection.
type Broker interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	Consume(ctx context.Context, queue string) (<-chan []byte, error)
	DeclareExchange(name, kind string) error
	DeclareQueue(name string) (string, error)
	BindQueue(queue, exchange, routingKey string) error
	Connect(ctx context.Context) error
	Close() error
}

// rmqBroker wraps a rabbitmq.Connection to implement the Broker interface.
type rmqBroker struct {
	conn *rabbitmq.Connection
}

func NewBroker(conn *rabbitmq.Connection) Broker {
	return &rmqBroker{conn: conn}
}

func (b *rmqBroker) Connect(ctx context.Context) error { return b.conn.Connect(ctx) }
func (b *rmqBroker) Close() error { return b.conn.Close() }
func (b *rmqBroker) DeclareExchange(name, kind string) error { return b.conn.DeclareExchange(name, kind) }
func (b *rmqBroker) DeclareQueue(name string) (string, error) { return b.conn.DeclareQueue(name) }
func (b *rmqBroker) BindQueue(queue, exchange, routingKey string) error { return b.conn.BindQueue(queue, exchange, routingKey) }
func (b *rmqBroker) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return b.conn.Publish(ctx, exchange, routingKey, body)
}

func (b *rmqBroker) Consume(ctx context.Context, queue string) (<-chan []byte, error) {
	deliveries, err := b.conn.Consume(ctx, queue)
	if err != nil {
		return nil, err
	}

	out := make(chan []byte)
	go func() {
		defer close(out)
		for d := range deliveries {
			out <- d.Body
		}
	}()
	return out, nil
}

// Producer publishes signaling messages to RabbitMQ.
type Producer struct {
	broker Broker
}

// NewProducer creates a new signal Producer.
func NewProducer(broker Broker) *Producer {
	return &Producer{broker: broker}
}

// Publish sends a signaling message on the voxmeet.signal exchange.
func (p *Producer) Publish(ctx context.Context, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal signal message: %w", err)
	}

	return p.broker.Publish(ctx,
		rabbitmq.ExchangeName("signal"),
		rabbitmq.RoutingKey("signal", msg.RoomID),
		body,
	)
}

// DispatchFunc is a callback for processing an incoming signaling message.
type DispatchFunc func(ctx context.Context, msg Message) error

// Consumer consumes signaling messages from RabbitMQ.
type Consumer struct {
	broker    Broker
	producer  *Producer
	dispatch  DispatchFunc
	queueName string
}

// NewConsumer creates a new signal Consumer.
func NewConsumer(broker Broker, producer *Producer) (*Consumer, error) {
	if broker == nil {
		return nil, fmt.Errorf("broker is required")
	}

	return &Consumer{
		broker:   broker,
		producer: producer,
		dispatch: nil,
	}, nil
}

// Start begins consuming signaling messages. Blocks until ctx is cancelled.
func (c *Consumer) Start(ctx context.Context) error {
	if err := c.broker.Connect(ctx); err != nil {
		return fmt.Errorf("connect to broker: %w", err)
	}
	defer c.broker.Close()

	if err := c.broker.DeclareExchange(rabbitmq.ExchangeName("signal"), "topic"); err != nil {
		return fmt.Errorf("declare signal exchange: %w", err)
	}

	queue, err := c.broker.DeclareQueue("")
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	if err := c.broker.BindQueue(queue, rabbitmq.ExchangeName("signal"), "signaling.room.*"); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	msgs, err := c.broker.Consume(ctx, queue)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	for msgBytes := range msgs {
		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		if c.dispatch != nil {
			if err := c.dispatch(ctx, msg); err != nil {
				// Log but continue processing
			}
		}
	}

	return nil
}

// WithDispatch sets the dispatch function for handling incoming messages.
func (c *Consumer) WithDispatch(fn DispatchFunc) *Consumer {
	c.dispatch = fn
	return c
}
