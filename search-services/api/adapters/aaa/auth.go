package aaa

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const secretKey = "something secret here" // token sign key
const adminRole = "superuser"             // token subject

// Authentication, Authorization, Accounting
type AAA struct {
	users    map[string]string
	tokenTTL time.Duration
	log      *slog.Logger
}

func New(tokenTTL time.Duration, log *slog.Logger) (AAA, error) {
	const adminUser = "ADMIN_USER"
	const adminPass = "ADMIN_PASSWORD"
	user, ok := os.LookupEnv(adminUser)
	if !ok {
		return AAA{}, fmt.Errorf("could not get admin user from environment")
	}
	password, ok := os.LookupEnv(adminPass)
	if !ok {
		return AAA{}, fmt.Errorf("could not get admin password from environment")
	}

	return AAA{
		users:    map[string]string{user: password},
		tokenTTL: tokenTTL,
		log:      log,
	}, nil
}

func (a AAA) Login(name, password string) (string, error) {
	pass, ok := a.users[name]
	if !ok || pass != password {
		return "", fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   adminRole,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		a.log.Error("failed to generate token", "error", err)
		return "", fmt.Errorf("failed to generate token")
	}

	return tokenString, nil
}

func (a AAA) Verify(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		a.log.Error("failed to parse token", "error", err)
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims")
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return fmt.Errorf("error get subject")
	}

	if subject != adminRole {
		return fmt.Errorf("invalid subject")
	}

	return nil
}
