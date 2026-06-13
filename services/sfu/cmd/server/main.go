package main

import (
	"context"
	"fmt"
	"log"
	ossignal "os/signal"
	"syscall"

	"github.com/awwal/voxmeet/pkgs/rabbitmq"
	"github.com/awwal/voxmeet/sfu/internal/config"
	"github.com/awwal/voxmeet/sfu/internal/room"
	"github.com/awwal/voxmeet/sfu/internal/signaling"
)

func main() {
	cfg := config.FromEnv()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	conn := rabbitmq.New(rabbitmq.Config{
		URL: cfg.RabbitMQURL,
	})

	broker := signaling.NewBroker(conn)
	producer := signaling.NewProducer(broker)
	manager := room.NewManager()

	iceServers := make([]string, 0, len(cfg.ICEServers))
	for _, s := range cfg.ICEServers {
		iceServers = append(iceServers, s.URLs...)
	}

	dispatcher := signaling.NewMediaDispatcher(manager, producer, iceServers)

	consumer, err := signaling.NewConsumer(broker, producer)
	if err != nil {
		log.Fatalf("create consumer: %v", err)
	}

	consumer.WithDispatch(dispatcher.DispatchFunc())

	ctx, stop := ossignal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("sfu starting...")
	if err := consumer.Start(ctx); err != nil {
		log.Printf("sfu stopped: %v", err)
	}
	dispatcher.CloseAll()
	manager.StopAll()
	fmt.Println("sfu stopped")
}
