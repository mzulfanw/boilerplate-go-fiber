package jwt

import (
	"testing"
	"time"
)

func TestManagerGenerateParse(t *testing.T) {
	manager, err := NewManager("secret", "issuer", time.Minute)
	if err != nil {
		t.Fatalf("expected manager, got error: %v", err)
	}

	token, expiresIn, err := manager.GenerateAccessToken("user-1", []string{"admin"}, []string{"user.read"}, 2)
	if err != nil {
		t.Fatalf("expected token, got error: %v", err)
	}
	if expiresIn <= 0 {
		t.Fatalf("expected positive expiresIn, got %d", expiresIn)
	}

	claims, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("expected claims, got error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
	if claims.Issuer != "issuer" {
		t.Fatalf("expected issuer issuer, got %s", claims.Issuer)
	}
	if claims.TokenVersion != 2 {
		t.Fatalf("expected token version 2, got %d", claims.TokenVersion)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Fatalf("expected role admin, got %v", claims.Roles)
	}
	if len(claims.Permissions) != 1 || claims.Permissions[0] != "user.read" {
		t.Fatalf("expected permission user.read, got %v", claims.Permissions)
	}
}

func TestManagerParseInvalidToken(t *testing.T) {
	manager, err := NewManager("secret", "issuer", time.Minute)
	if err != nil {
		t.Fatalf("expected manager, got error: %v", err)
	}

	if _, err := manager.ParseAccessToken("not-a-token"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}
