package rbac

import (
	"context"
	"errors"
	"sort"
	"strings"

	rbacdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/rbac"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("rbac: repository is nil")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) ListRoles(ctx context.Context, filter rbacdomain.ListFilterRole) (rbacdomain.ListRole, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	return s.repo.ListRoles(ctx, filter)
}

func (s *Service) GetRole(ctx context.Context, id string) (rbacdomain.Role, error) {
	if strings.TrimSpace(id) == "" {
		return rbacdomain.Role{}, rbacdomain.ErrInvalidInput
	}
	return s.repo.GetRole(ctx, id)
}

func (s *Service) CreateRole(ctx context.Context, name, description string) (rbacdomain.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return rbacdomain.Role{}, rbacdomain.ErrInvalidInput
	}
	return s.repo.CreateRole(ctx, name, strings.TrimSpace(description))
}

func (s *Service) UpdateRole(ctx context.Context, id, name, description string) (rbacdomain.Role, error) {
	if strings.TrimSpace(id) == "" {
		return rbacdomain.Role{}, rbacdomain.ErrInvalidInput
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return rbacdomain.Role{}, rbacdomain.ErrInvalidInput
	}
	return s.repo.UpdateRole(ctx, id, name, strings.TrimSpace(description))
}

func (s *Service) DeleteRole(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return rbacdomain.ErrInvalidInput
	}
	return s.repo.DeleteRole(ctx, id)
}

func (s *Service) ListPermissions(ctx context.Context, filter rbacdomain.ListFilterPermission) (rbacdomain.ListPermission, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	return s.repo.ListPermissions(ctx, filter)
}

func (s *Service) GetPermission(ctx context.Context, id string) (rbacdomain.Permission, error) {
	if strings.TrimSpace(id) == "" {
		return rbacdomain.Permission{}, rbacdomain.ErrInvalidInput
	}
	return s.repo.GetPermission(ctx, id)
}

func (s *Service) CreatePermission(ctx context.Context, name, description string) (rbacdomain.Permission, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return rbacdomain.Permission{}, rbacdomain.ErrInvalidInput
	}
	return s.repo.CreatePermission(ctx, name, strings.TrimSpace(description))
}

func (s *Service) UpdatePermission(ctx context.Context, id, name, description string) (rbacdomain.Permission, error) {
	if strings.TrimSpace(id) == "" {
		return rbacdomain.Permission{}, rbacdomain.ErrInvalidInput
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return rbacdomain.Permission{}, rbacdomain.ErrInvalidInput
	}
	return s.repo.UpdatePermission(ctx, id, name, strings.TrimSpace(description))
}

func (s *Service) DeletePermission(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return rbacdomain.ErrInvalidInput
	}
	return s.repo.DeletePermission(ctx, id)
}

func (s *Service) ListRolePermissions(ctx context.Context, roleID string) ([]rbacdomain.Permission, error) {
	if strings.TrimSpace(roleID) == "" {
		return nil, rbacdomain.ErrInvalidInput
	}
	return s.repo.ListRolePermissions(ctx, roleID)
}

func (s *Service) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	if strings.TrimSpace(roleID) == "" {
		return rbacdomain.ErrInvalidInput
	}
	normalized := normalizeIDs(permissionIDs)
	return s.repo.ReplaceRolePermissions(ctx, roleID, normalized)
}

func normalizeIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		seen[trimmed] = struct{}{}
	}
	if len(seen) == 0 {
		return nil
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}
