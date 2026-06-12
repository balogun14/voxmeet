package auth_test

import (
	"testing"
	"time"

	"github.com/awwal/voxmeet/api-gateway/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_JWT_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-key"
	userID := uuid.New().String()

	token, expiresAt, err := auth.GenerateJWT(secret, userID, 1*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	assert.False(t, expiresAt.IsZero())
	assert.True(t, expiresAt.After(time.Now()))

	claims, err := auth.ValidateJWT(secret, token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func Test_JWT_InvalidSecret(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New().String()

	token, _, err := auth.GenerateJWT(secret, userID, 1*time.Hour)
	require.NoError(t, err)

	_, err = auth.ValidateJWT("wrong-secret", token)
	assert.Error(t, err)
}

func Test_JWT_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New().String()

	token, _, err := auth.GenerateJWT(secret, userID, -1*time.Hour)
	require.NoError(t, err)

	_, err = auth.ValidateJWT(secret, token)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
	assert.False(t, auth.IsValidJWT(secret, token))
}

func Test_JWT_InvalidFormat(t *testing.T) {
	_, err := auth.ValidateJWT("secret", "not-a-valid-token")
	assert.Error(t, err)
	assert.False(t, auth.IsValidJWT("secret", "not-a-valid-token"))
}

func Test_JWT_IsValidJWT(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New().String()

	token, _, err := auth.GenerateJWT(secret, userID, 1*time.Hour)
	require.NoError(t, err)

	assert.True(t, auth.IsValidJWT(secret, token))
	assert.False(t, auth.IsValidJWT(secret, "bad-token"))
	assert.False(t, auth.IsValidJWT("wrong-secret", token))
}

func Test_Password_HashAndCompare(t *testing.T) {
	password := "my-secure-password-123!"

	hash, err := auth.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	assert.True(t, auth.CheckPassword(hash, password))
	assert.False(t, auth.CheckPassword(hash, "wrong-password"))
	assert.False(t, auth.CheckPassword(hash, ""))
}

func Test_Password_NoTwoHashesEqual(t *testing.T) {
	password := "same-password"

	hash1, err := auth.HashPassword(password)
	require.NoError(t, err)

	hash2, err := auth.HashPassword(password)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
	assert.True(t, auth.CheckPassword(hash1, password))
	assert.True(t, auth.CheckPassword(hash2, password))
}

func Test_Password_EmptyString(t *testing.T) {
	_, err := auth.HashPassword("")
	assert.Error(t, err)
}

func Test_Password_InvalidHash(t *testing.T) {
	assert.False(t, auth.CheckPassword("not-a-bcrypt-hash", "password"))
}
