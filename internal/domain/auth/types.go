package auth

import "time"

type User struct {
	ID                  string
	Email               string
	PasswordHash        string
	IsActive            bool
	FailedLoginAttempts int
	LockedUntil         *time.Time
	TokenVersion        int
}

type RefreshToken struct {
	ID             string
	UserID         string
	TokenHash      string
	ReplacedByHash string
	ExpiresAt      time.Time
	RevokedAt      *time.Time
	CreatedAt      time.Time
	IPAddress      string
	UserAgent      string
}

type AuthState struct {
	IsActive     bool
	TokenVersion int
}
