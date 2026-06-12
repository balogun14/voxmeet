package config_test

import (
	"testing"

	"github.com/awwal/voxmeet/sfu/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_Config_Defaults(t *testing.T) {
	cfg := config.Default()
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQURL)
	assert.Equal(t, uint16(50000), cfg.WebRTCPortMin)
	assert.Equal(t, uint16(50100), cfg.WebRTCPortMax)
	assert.Contains(t, cfg.ICEServers[0].URLs[0], "stun:stun.l.google.com")
}

func Test_Config_Validate(t *testing.T) {
	t.Run("valid config passes", func(t *testing.T) {
		cfg := config.Default()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("empty rabbitmq URL fails", func(t *testing.T) {
		cfg := config.Config{RabbitMQURL: ""}
		assert.Error(t, cfg.Validate())
	})

	t.Run("invalid port range fails", func(t *testing.T) {
		cfg := config.Config{RabbitMQURL: "amqp://localhost:5672/", WebRTCPortMin: 100, WebRTCPortMax: 50}
		assert.Error(t, cfg.Validate())
	})
}
