package config_test

import (
	"os"
	"testing"

	"github.com/awwal/voxmeet/chat-service/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_ChatConfig_Defaults(t *testing.T) {
	cfg := config.Default()
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQURL)
	assert.NotEmpty(t, cfg.DatabaseURL)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func Test_ChatConfig_FromEnv(t *testing.T) {
	os.Setenv("RABBITMQ_URL", "amqp://admin:admin@rabbit:5672/")
	os.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/voxmeet")
	defer os.Clearenv()

	cfg := config.FromEnv()
	assert.Equal(t, "amqp://admin:admin@rabbit:5672/", cfg.RabbitMQURL)
	assert.Equal(t, "postgres://user:pass@db:5432/voxmeet", cfg.DatabaseURL)
}

func Test_ChatConfig_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := config.Config{RabbitMQURL: "amqp://localhost/", DatabaseURL: "postgres://localhost/voxmeet"}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("missing rabbitmq", func(t *testing.T) {
		cfg := config.Config{DatabaseURL: "postgres://localhost/voxmeet"}
		assert.Error(t, cfg.Validate())
	})
	t.Run("missing database", func(t *testing.T) {
		cfg := config.Config{RabbitMQURL: "amqp://localhost/"}
		assert.Error(t, cfg.Validate())
	})
}
