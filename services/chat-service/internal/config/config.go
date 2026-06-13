package config

import (
	"errors"
	"os"
)

type Config struct {
	RabbitMQURL string
	DatabaseURL string
	LogLevel    string
}

func Default() Config {
	return Config{
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/voxmeet?sslmode=disable",
		LogLevel:    "debug",
	}
}

func FromEnv() Config {
	return Config{
		RabbitMQURL: envOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/voxmeet?sslmode=disable"),
		LogLevel:    envOrDefault("LOG_LEVEL", "debug"),
	}
}

func (c *Config) Validate() error {
	if c.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	return nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
