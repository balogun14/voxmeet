package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	ossignal "os/signal"
	"syscall"

	"github.com/awwal/voxmeet/pkgs/rabbitmq"
	"github.com/awwal/voxmeet/room-service/internal/config"
	"github.com/awwal/voxmeet/room-service/internal/handler"
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

	rmq.DeclareExchange(rabbitmq.ExchangeName("rpc"), "direct")
	q, _ := rmq.DeclareQueue("voxmeet.rpc.room-service")
	rmq.BindQueue(q, rabbitmq.ExchangeName("rpc"), "rpc.room")

	msgs, err := rmq.Consume(context.Background(), q)
	if err != nil {
		log.Fatalf("consume: %v", err)
	}

	h := handler.NewHandler(nil)

	sigCh := make(chan os.Signal, 1)
	ossignal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("room-service starting...")

	go func() {
		<-sigCh
		rmq.Close()
	}()

	for d := range msgs {
		var req struct {
			Action string          `json:"action"`
			Data   json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(d.Body, &req); err != nil {
			continue
		}
		_ = h
		_ = req
	}

	fmt.Println("room-service stopped")
}
