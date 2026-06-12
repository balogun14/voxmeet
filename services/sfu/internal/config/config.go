package config

import (
	"errors"
	"os"
)

type ICEServerConfig struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type Config struct {
	RabbitMQURL     string
	WebRTCPortMin   uint16
	WebRTCPortMax   uint16
	ICEServers      []ICEServerConfig
	PublicIP        string
}

func Default() Config {
	return Config{
		RabbitMQURL:   "amqp://guest:guest@localhost:5672/",
		WebRTCPortMin: 50000,
		WebRTCPortMax: 50100,
		ICEServers: []ICEServerConfig{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
}

func FromEnv() Config {
	cfg := Default()
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		cfg.RabbitMQURL = v
	}
	cfg.PublicIP = os.Getenv("PUBLIC_IP")
	return cfg
}

func (c *Config) Validate() error {
	if c.RabbitMQURL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	if c.WebRTCPortMax <= c.WebRTCPortMin {
		return errors.New("WebRTCPortMax must be greater than WebRTCPortMin")
	}
	return nil
}
