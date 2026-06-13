package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	ossignal "os/signal"
	"syscall"
	"time"

	"github.com/awwal/voxmeet/pkgs/rabbitmq"
	"github.com/awwal/voxmeet/presence-service/internal/config"
	"github.com/awwal/voxmeet/presence-service/internal/presence"
)

func main() {
	cfg := config.FromEnv()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	ctx := context.Background()
	rmq := rabbitmq.New(rabbitmq.Config{URL: cfg.RabbitMQURL})
	if err := rmq.Connect(ctx); err != nil {
		log.Fatalf("connect to rabbitmq: %v", err)
	}
	defer rmq.Close()

	rmq.DeclareExchange(rabbitmq.ExchangeName("presence"), "topic")

	q, _ := rmq.DeclareQueue("voxmeet.presence-service.events")
	rmq.BindQueue(q, rabbitmq.ExchangeName("presence"), "presence.room.*")

	msgs, err := rmq.Consume(ctx, q)
	if err != nil {
		log.Fatalf("consume messages: %v", err)
	}

	tr := presence.NewTracker(time.Duration(cfg.HeartbeatTimeout) * time.Second)
	defer tr.Stop()

	sigCh := make(chan os.Signal, 1)
	ossignal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("presence-service starting...")

	go func() {
		<-sigCh
		fmt.Println("\nshutting down...")
		rmq.Close()
		tr.Stop()
	}()

	for d := range msgs {
		var evt struct {
			Action   string `json:"action"`
			UserID   string `json:"user_id"`
			RoomID   string `json:"room_id"`
			Typing   *bool  `json:"typing,omitempty"`
		}
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			continue
		}

		switch evt.Action {
		case "online":
			tr.UserOnline(evt.UserID, evt.RoomID)
		case "offline":
			tr.UserOffline(evt.UserID)
		case "heartbeat":
			tr.Heartbeat(evt.UserID)
		case "typing":
			if evt.Typing != nil {
				tr.SetTyping(evt.UserID, *evt.Typing)
			}
		}
	}

	fmt.Println("presence-service stopped")
}
