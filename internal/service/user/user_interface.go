package user

import (
	"context"

	rbacdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/rbac"
	userdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/user"
)

type Repository interface {
	ListUsers(ctx context.Context, filter userdomain.ListFilter) (userdomain.ListResult, error)
	GetUser(ctx context.Context, id string) (userdomain.User, error)
	CreateUser(ctx context.Context, email, passwordHash string, isActive bool, roleIDs []string) (userdomain.User, error)
	UpdateUser(ctx context.Context, id, email, passwordHash string, isActive bool, bumpTokenVersion bool) (userdomain.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUserRoles(ctx context.Context, userID string) ([]rbacdomain.Role, error)
	ReplaceUserRoles(ctx context.Context, userID string, roleIDs []string) error
}
