package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("недействительный токен")
	ErrExpiredToken = errors.New("срок действия токена истек")
)

type TokenManager interface {
	Generate(userID string, ttl time.Duration) (string, error)
	Validate(tokenStr string) (string, error)
}

type JWTManager struct {
	secretKey []byte
}

func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
	}
}

func (m *JWTManager) Generate(userID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) Validate(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неизвестный метод подписи: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return sub, nil
}
