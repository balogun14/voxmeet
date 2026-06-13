package config_test

import (
	"os"
	"testing"

	"github.com/awwal/voxmeet/presence-service/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_PresenceConfig_Defaults(t *testing.T) {
	cfg := config.Default()
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQURL)
	assert.Equal(t, 30, cfg.HeartbeatTimeout)
}

func Test_PresenceConfig_FromEnv(t *testing.T) {
	os.Setenv("RABBITMQ_URL", "amqp://admin:admin@rabbit:5672/")
	os.Setenv("HEARTBEAT_TIMEOUT", "60")
	defer os.Clearenv()

	cfg := config.FromEnv()
	assert.Equal(t, "amqp://admin:admin@rabbit:5672/", cfg.RabbitMQURL)
	assert.Equal(t, 60, cfg.HeartbeatTimeout)
}

func Test_PresenceConfig_Validate(t *testing.T) {
	cfg := config.Config{RabbitMQURL: "amqp://localhost/", HeartbeatTimeout: 30}
	assert.NoError(t, cfg.Validate())

	cfg = config.Config{RabbitMQURL: "amqp://localhost/"}
	assert.Error(t, cfg.Validate())

	cfg = config.Config{HeartbeatTimeout: 30}
	assert.Error(t, cfg.Validate())
}
