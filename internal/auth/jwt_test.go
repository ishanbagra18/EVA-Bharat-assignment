package auth_test

import (
	"testing"

	"ticket-system/internal/auth"
)

func TestPasswordHashing(t *testing.T) {
	password := "secret123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !auth.CheckPasswordHash(password, hash) {
		t.Errorf("CheckPasswordHash failed for correct password")
	}

	if auth.CheckPasswordHash("wrongpassword", hash) {
		t.Errorf("CheckPasswordHash passed for incorrect password")
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	userID := "usr-12345"
	token, err := auth.GenerateToken(userID, 1)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate valid token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}
}

func TestInvalidJWT(t *testing.T) {
	_, err := auth.ValidateToken("invalid.jwt.token")
	if err == nil {
		t.Errorf("expected error validating malformed JWT, got nil")
	}
}
