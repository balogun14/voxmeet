package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/awwal/voxmeet/api-gateway/internal/auth"
	"github.com/awwal/voxmeet/api-gateway/internal/config"
	"github.com/awwal/voxmeet/api-gateway/internal/db"
	"github.com/awwal/voxmeet/api-gateway/internal/docs"
	"github.com/awwal/voxmeet/api-gateway/internal/handler"
	"github.com/awwal/voxmeet/api-gateway/internal/middleware"
	"github.com/awwal/voxmeet/api-gateway/internal/migrate"
	"github.com/awwal/voxmeet/api-gateway/internal/ws"
	"github.com/awwal/voxmeet/pkgs/rabbitmq"
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

	queries := db.New(pool)

	// Run migrations
	if err := migrate.Run(context.Background(), pool); err != nil {
		log.Printf("migration warning: %v", err)
	}

	// Connect to RabbitMQ
	rmq := rabbitmq.New(rabbitmq.Config{URL: cfg.RabbitMQURL})
	if err := rmq.Connect(context.Background()); err != nil {
		log.Fatalf("connect to rabbitmq: %v", err)
	}
	defer rmq.Close()

	// WebSocket hub
	hub := ws.NewHub()
	hub.SetPublisher(func(exchange, routingKey string, body []byte) {
		if err := rmq.Publish(context.Background(), exchange, routingKey, body); err != nil {
			log.Printf("publish to rmq: %v", err)
		}
	})
	go hub.Run()

	// Consume signal messages from RabbitMQ and route to WS clients
	rmq.DeclareExchange(rabbitmq.ExchangeName("signal"), "topic")
	q, err := rmq.DeclareQueue("voxmeet.api-gateway.signal")
	if err != nil {
		log.Fatalf("declare queue: %v", err)
	}
	if err := rmq.BindQueue(q, rabbitmq.ExchangeName("signal"), "signal.room.*"); err != nil {
		log.Fatalf("bind queue: %v", err)
	}
	sigMsgs, err := rmq.Consume(context.Background(), q)
	if err != nil {
		log.Fatalf("consume queue: %v", err)
	}
	go hub.ConsumeSignalMessages(context.Background(), sigMsgs)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /api/v1/health", handler.Health)

	authH := handler.NewAuthHandler(queries, cfg.JWTSecret, 24*time.Hour)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)

	// API docs
	docsMux := docs.Handler()
	mux.Handle("GET /api/v1/docs", docsMux)
	mux.Handle("GET /api/v1/docs/", docsMux)

	// WebSocket
	mux.Handle("GET /api/v1/ws", ws.Handler(hub, cfg.JWTSecret))

	// Protected routes
	roomH := handler.NewRoomHandler(queries)
	mux.Handle("GET /api/v1/rooms/{id}", auth.Middleware(cfg.JWTSecret, http.HandlerFunc(roomH.GetByID)))
	mux.Handle("PUT /api/v1/rooms/{id}", auth.Middleware(cfg.JWTSecret, http.HandlerFunc(roomH.Update)))
	mux.Handle("DELETE /api/v1/rooms/{id}", auth.Middleware(cfg.JWTSecret, http.HandlerFunc(roomH.Delete)))
	mux.Handle("GET /api/v1/rooms", auth.Middleware(cfg.JWTSecret, http.HandlerFunc(roomH.List)))
	mux.Handle("POST /api/v1/rooms", auth.Middleware(cfg.JWTSecret, http.HandlerFunc(roomH.Create)))
	mux.Handle("GET /api/v1/me", auth.Middleware(cfg.JWTSecret, http.HandlerFunc(handler.CurrentUser(queries))))

	srv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      middleware.CORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("api-gateway listening on %s\n", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nshutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	fmt.Println("server stopped")
}
