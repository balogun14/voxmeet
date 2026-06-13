package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	ossignal "os/signal"
	"syscall"

	"github.com/awwal/voxmeet/chat-service/internal/config"
	"github.com/awwal/voxmeet/chat-service/internal/handler"
	"github.com/awwal/voxmeet/pkgs/rabbitmq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rpcPublisher struct {
	conn *rabbitmq.Connection
}

func (p *rpcPublisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return p.conn.Publish(ctx, exchange, routingKey, body)
}

func main() {
	cfg := config.FromEnv()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	rmq := rabbitmq.New(rabbitmq.Config{URL: cfg.RabbitMQURL})
	if err := rmq.Connect(context.Background()); err != nil {
		log.Fatalf("connect to rabbitmq: %v", err)
	}
	defer rmq.Close()

	rmq.DeclareExchange(rabbitmq.ExchangeName("chat"), "topic")
	q, _ := rmq.DeclareQueue("voxmeet.chat-service.messages")
	rmq.BindQueue(q, rabbitmq.ExchangeName("chat"), "chat.room.*")

	msgs, err := rmq.Consume(context.Background(), q)
	if err != nil {
		log.Fatalf("consume messages: %v", err)
	}

	h := handler.NewHandler(pool, &rpcPublisher{conn: rmq})

	sigCh := make(chan os.Signal, 1)
	ossignal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("chat-service starting...")

	go func() {
		<-sigCh
		fmt.Println("\nshutting down...")
		rmq.Close()
		pool.Close()
	}()

	for d := range msgs {
		var evt struct {
			Action  string `json:"action"`
			RoomID  string `json:"room_id"`
			UserID  string `json:"user_id"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			continue
		}
		switch evt.Action {
		case "chat.send":
			if err := h.HandleMessage(context.Background(), handler.MessageEvent{
				RoomID: evt.RoomID, UserID: evt.UserID, Content: evt.Content,
			}); err != nil {
				log.Printf("handle message: %v", err)
			}
		}
	}

	fmt.Println("chat-service stopped")
}
