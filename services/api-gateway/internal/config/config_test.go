package config_test

import (
	"os"
	"testing"

	"github.com/awwal/voxmeet/api-gateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Config_Defaults(t *testing.T) {
	cfg := config.Default()
	assert.Equal(t, ":8080", cfg.HTTPPort)
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQURL)
	assert.Equal(t, "postgres://postgres:postgres@localhost:5432/voxmeet?sslmode=disable", cfg.DatabaseURL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "", cfg.JWTSecret)
}

func Test_Config_FromEnv(t *testing.T) {
	os.Setenv("HTTP_PORT", ":9090")
	os.Setenv("RABBITMQ_URL", "amqp://admin:admin@rabbit:5672/")
	os.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/voxmeet")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("JWT_SECRET", "super-secret-key")
	defer os.Clearenv()

	cfg := config.FromEnv()
	assert.Equal(t, ":9090", cfg.HTTPPort)
	assert.Equal(t, "amqp://admin:admin@rabbit:5672/", cfg.RabbitMQURL)
	assert.Equal(t, "postgres://user:pass@db:5432/voxmeet", cfg.DatabaseURL)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "super-secret-key", cfg.JWTSecret)
}

func Test_Config_Validate(t *testing.T) {
	t.Run("valid config passes", func(t *testing.T) {
		cfg := config.Config{
			HTTPPort:     ":8080",
			RabbitMQURL:  "amqp://guest:guest@localhost:5672/",
			DatabaseURL:  "postgres://postgres:postgres@localhost:5432/voxmeet?sslmode=disable",
			JWTSecret:    "some-secret",
			LogLevel:     "debug",
		}
		err := cfg.Validate()
		require.NoError(t, err)
	})

	t.Run("missing JWT secret fails", func(t *testing.T) {
		cfg := config.Config{HTTPPort: ":8080", RabbitMQURL: "amqp://localhost/", DatabaseURL: "postgres://localhost/voxmeet"}
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("missing port fails", func(t *testing.T) {
		cfg := config.Config{RabbitMQURL: "amqp://localhost/", DatabaseURL: "postgres://localhost/voxmeet", JWTSecret: "secret"}
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("missing rabbitmq URL fails", func(t *testing.T) {
		cfg := config.Config{HTTPPort: ":8080", DatabaseURL: "postgres://localhost/voxmeet", JWTSecret: "secret"}
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("missing database URL fails", func(t *testing.T) {
		cfg := config.Config{HTTPPort: ":8080", RabbitMQURL: "amqp://localhost/", JWTSecret: "secret"}
		err := cfg.Validate()
		assert.Error(t, err)
	})
}
