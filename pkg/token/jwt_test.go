package token

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	manager := NewJWTManager("test_secret_key_12345")
	userID := "user-1234-uuid"

	tokenStr, err := manager.Generate(userID, 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	parsedID, err := manager.Validate(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if parsedID != userID {
		t.Fatalf("expected userID %q, got %q", userID, parsedID)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	manager := NewJWTManager("test_secret_key_12345")
	userID := "user-1234-uuid"

	tokenStr, err := manager.Generate(userID, -1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	_, err = manager.Validate(tokenStr)
	if err != ErrExpiredToken {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
}
