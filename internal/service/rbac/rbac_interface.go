package rbac

import (
	"context"

	rbacdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/rbac"
)

type Repository interface {
	ListRoles(ctx context.Context, filter rbacdomain.ListFilterRole) (rbacdomain.ListRole, error)
	GetRole(ctx context.Context, id string) (rbacdomain.Role, error)
	CreateRole(ctx context.Context, name, description string) (rbacdomain.Role, error)
	UpdateRole(ctx context.Context, id, name, description string) (rbacdomain.Role, error)
	DeleteRole(ctx context.Context, id string) error

	ListPermissions(ctx context.Context, filter rbacdomain.ListFilterPermission) (rbacdomain.ListPermission, error)
	GetPermission(ctx context.Context, id string) (rbacdomain.Permission, error)
	CreatePermission(ctx context.Context, name, description string) (rbacdomain.Permission, error)
	UpdatePermission(ctx context.Context, id, name, description string) (rbacdomain.Permission, error)
	DeletePermission(ctx context.Context, id string) error

	ListRolePermissions(ctx context.Context, roleID string) ([]rbacdomain.Permission, error)
	ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error
}