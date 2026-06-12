package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPPort    string
	RabbitMQURL string
	DatabaseURL string
	JWTSecret   string
	LogLevel    string
}

func Default() Config {
	return Config{
		HTTPPort:    ":8080",
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/voxmeet?sslmode=disable",
		LogLevel:    "debug",
	}
}

func FromEnv() Config {
	return Config{
		HTTPPort:    envOrDefault("HTTP_PORT", ":8080"),
		RabbitMQURL: envOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/voxmeet?sslmode=disable"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		LogLevel:    envOrDefault("LOG_LEVEL", "debug"),
	}
}

func (c *Config) Validate() error {
	if c.HTTPPort == "" {
		return errors.New("HTTP_PORT is required")
	}
	if c.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	return nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
