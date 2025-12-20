package auth

import (
	"context"

	authdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/auth"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (authdomain.User, error)
	FindUserByID(ctx context.Context, id string) (authdomain.User, error)
	GetUserAuthState(ctx context.Context, id string) (authdomain.AuthState, error)
	ListUserRoles(ctx context.Context, userID string) ([]string, error)
	ListUserPermissions(ctx context.Context, userID string) ([]string, error)
	CreateRefreshToken(ctx context.Context, token authdomain.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (authdomain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash, replacedByHash string) error
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
	RecordLoginFailure(ctx context.Context, userID string, maxAttempts int, lockoutSeconds int64) error
	ResetLoginFailures(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}
