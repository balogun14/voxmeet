package config_test

import (
	"testing"

	"github.com/awwal/voxmeet/room-service/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_RoomConfig_Defaults(t *testing.T) {
	cfg := config.Default()
	assert.NotEmpty(t, cfg.RabbitMQURL)
	assert.NotEmpty(t, cfg.DatabaseURL)
}

func Test_RoomConfig_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		assert.NoError(t, (&config.Config{RabbitMQURL: "amqp://localhost/", DatabaseURL: "postgres://localhost/voxmeet"}).Validate())
	})
	t.Run("missing rabbitmq", func(t *testing.T) {
		assert.Error(t, (&config.Config{DatabaseURL: "postgres://localhost/voxmeet"}).Validate())
	})
}
