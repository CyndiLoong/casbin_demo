package tests

import (
	"testing"

	"casbin-demo/internal/config"
	jwtpkg "casbin-demo/pkg/jwt"
)

func init() {
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret-key",
			ExpireHours: 24,
			Issuer:      "test",
		},
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	token, err := jwtpkg.GenerateToken(1, "test-uuid", "admin", []string{"admin"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	claims, err := jwtpkg.ParseToken(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", claims.UserID)
	}

	if claims.Username != "admin" {
		t.Errorf("Expected username 'admin', got %s", claims.Username)
	}

	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Errorf("Expected roles [admin], got %v", claims.Roles)
	}
}

func TestParseInvalidToken(t *testing.T) {
	_, err := jwtpkg.ParseToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestParseEmptyToken(t *testing.T) {
	_, err := jwtpkg.ParseToken("")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}
