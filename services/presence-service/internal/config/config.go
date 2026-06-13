package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	RabbitMQURL       string
	HeartbeatTimeout  int // seconds
	LogLevel          string
}

func Default() Config {
	return Config{
		RabbitMQURL:      "amqp://guest:guest@localhost:5672/",
		HeartbeatTimeout: 30,
		LogLevel:         "debug",
	}
}

func FromEnv() Config {
	return Config{
		RabbitMQURL:      envOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		HeartbeatTimeout: envOrDefaultInt("HEARTBEAT_TIMEOUT", 30),
		LogLevel:         envOrDefault("LOG_LEVEL", "debug"),
	}
}

func (c *Config) Validate() error {
	if c.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	if c.HeartbeatTimeout <= 0 {
		return errors.New("HEARTBEAT_TIMEOUT must be positive")
	}
	return nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envOrDefaultInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		i := 0
		_, _ = fmt.Sscanf(v, "%d", &i)
		if i > 0 {
			return i
		}
	}
	return defaultVal
}
