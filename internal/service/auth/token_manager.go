package auth

import "time"

type AccessClaims struct {
	Subject      string
	Issuer       string
	ExpiresAt    time.Time
	Roles        []string
	Permissions  []string
	TokenVersion int
}

type TokenManager interface {
	GenerateAccessToken(userID string, roles, permissions []string, tokenVersion int) (string, int64, error)
	ParseAccessToken(tokenString string) (AccessClaims, error)
}
