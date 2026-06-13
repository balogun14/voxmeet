package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	ossignal "os/signal"
	"syscall"

	"github.com/awwal/voxmeet/chat-service/internal/config"
	"github.com/awwal/voxmeet/chat-service/internal/handler"
	"github.com/awwal/voxmeet/pkgs/rabbitmq"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	rmq.DeclareExchange(rabbitmq.ExchangeName("rpc"), "direct")

	// Declare and bind queue for chat messages
	q, _ := rmq.DeclareQueue("voxmeet.chat-service.messages")
	rmq.BindQueue(q, rabbitmq.ExchangeName("chat"), "chat.room.*")

	msgs, err := rmq.Consume(context.Background(), q)
	if err != nil {
		log.Fatalf("consume messages: %v", err)
	}

	h := handler.NewHandler(nil, nil) // TODO: wire real store and publisher

	ctx, stop := ossignal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("chat-service starting...")

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
			h.HandleMessage(ctx, handler.MessageEvent{
				RoomID:  evt.RoomID,
				UserID:  evt.UserID,
				Content: evt.Content,
			})
		}
	}

	fmt.Println("chat-service stopped")
}
