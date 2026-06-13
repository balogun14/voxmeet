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
	"github.com/awwal/voxmeet/room-service/internal/store"
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

	queries := store.New(pool)
	h := handler.NewHandler(queries)

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

	sigCh := make(chan os.Signal, 1)
	ossignal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("room-service starting...")

	go func() {
		<-sigCh
		rmq.Close()
		pool.Close()
	}()

	for d := range msgs {
		var req struct {
			Action string          `json:"action"`
			Data   json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(d.Body, &req); err != nil {
			continue
		}

		respond := func(payload interface{}) {
			body, _ := json.Marshal(payload)
			rmq.Publish(context.Background(), rabbitmq.ExchangeName("rpc"), d.ReplyTo, body)
		}

		switch req.Action {
		case "create_room":
			var args struct {
				OwnerID string `json:"owner_id"`
				Name    string `json:"name"`
				Public  bool   `json:"is_public"`
				Max     int32  `json:"max_participants"`
			}
			json.Unmarshal(req.Data, &args)
			room, err := h.CreateRoom(context.Background(), args.OwnerID, args.Name, args.Public, args.Max)
			if err != nil {
				respond(map[string]string{"error": err.Error()})
			} else {
				respond(room)
			}
		case "get_room":
			var args struct {
				RoomID string `json:"room_id"`
			}
			json.Unmarshal(req.Data, &args)
			room, err := h.GetRoom(context.Background(), args.RoomID)
			if err != nil {
				respond(map[string]string{"error": err.Error()})
			} else {
				respond(room)
			}
		case "delete_room":
			var args struct {
				RoomID string `json:"room_id"`
			}
			json.Unmarshal(req.Data, &args)
			if err := h.DeleteRoom(context.Background(), args.RoomID); err != nil {
				respond(map[string]string{"error": err.Error()})
			} else {
				respond(map[string]string{"status": "deleted"})
			}
		case "add_member":
			var args struct {
				RoomID string `json:"room_id"`
				UserID string `json:"user_id"`
				Role   string `json:"role"`
			}
			json.Unmarshal(req.Data, &args)
			if err := h.AddMember(context.Background(), args.RoomID, args.UserID, args.Role); err != nil {
				respond(map[string]string{"error": err.Error()})
			} else {
				respond(map[string]string{"status": "ok"})
			}
		}
	}

	fmt.Println("room-service stopped")
}
