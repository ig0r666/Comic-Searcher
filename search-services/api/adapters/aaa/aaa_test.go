package aaa

import (
	"log/slog"
	"os"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		os.Setenv("ADMIN_USER", "admin")
		os.Setenv("ADMIN_PASSWORD", "password")
		defer os.Unsetenv("ADMIN_USER")
		defer os.Unsetenv("ADMIN_PASSWORD")

		aaa, err := New(time.Hour, slog.Default())
		assert.NoError(t, err)
		assert.Equal(t, "password", aaa.users["admin"])
	})

	t.Run("missing username", func(t *testing.T) {
		os.Unsetenv("ADMIN_USER")
		os.Setenv("ADMIN_PASSWORD", "password")
		defer os.Unsetenv("ADMIN_PASSWORD")

		_, err := New(time.Hour, slog.Default())
		assert.Error(t, err)
	})

	t.Run("missing password", func(t *testing.T) {
		os.Setenv("ADMIN_USER", "admin")
		os.Unsetenv("ADMIN_PASSWORD")
		defer os.Unsetenv("ADMIN_USER")

		_, err := New(time.Hour, slog.Default())
		assert.Error(t, err)
	})
}

func TestLogin(t *testing.T) {
	os.Setenv("ADMIN_USER", "admin")
	os.Setenv("ADMIN_PASSWORD", "password")
	defer os.Unsetenv("ADMIN_USER")
	defer os.Unsetenv("ADMIN_PASSWORD")

	aaa, _ := New(time.Hour, slog.Default())

	t.Run("successful login", func(t *testing.T) {
		token, err := aaa.Login("admin", "password")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("wrong username", func(t *testing.T) {
		_, err := aaa.Login("wrong", "password")
		assert.Error(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := aaa.Login("admin", "wrongpassword")
		assert.Error(t, err)
	})
}

func TestVerify(t *testing.T) {
	os.Setenv("ADMIN_USER", "admin")
	os.Setenv("ADMIN_PASSWORD", "password")
	defer os.Unsetenv("ADMIN_USER")
	defer os.Unsetenv("ADMIN_PASSWORD")

	aaa, _ := New(time.Hour, slog.Default())
	validToken, _ := aaa.Login("admin", "password")

	t.Run("valid token", func(t *testing.T) {
		err := aaa.Verify(validToken)
		assert.NoError(t, err)
	})

	t.Run("invalid token", func(t *testing.T) {
		err := aaa.Verify("invalidtoken")
		assert.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		aaa.tokenTTL = -time.Hour
		expiredToken, _ := aaa.Login("admin", "pass")
		err := aaa.Verify(expiredToken)
		assert.Error(t, err)
	})

	t.Run("invalid subject", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject: "test",
		})
		invalidToken, _ := token.SignedString([]byte(secretKey))
		err := aaa.Verify(invalidToken)
		assert.Error(t, err)
	})
}
