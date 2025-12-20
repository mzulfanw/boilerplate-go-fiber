package user

import (
	"context"
	"errors"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"

	rbacdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/rbac"
	userdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/user"
)

const passwordHashCost = 12

type Service struct {
	repo Repository
}

func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("user: repository is nil")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]userdomain.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *Service) GetUser(ctx context.Context, id string) (userdomain.User, error) {
	if strings.TrimSpace(id) == "" {
		return userdomain.User{}, userdomain.ErrInvalidInput
	}
	return s.repo.GetUser(ctx, id)
}

func (s *Service) CreateUser(ctx context.Context, email, password string, isActive *bool, roleIDs []string) (userdomain.User, error) {
	normalizedEmail := normalizeEmail(email)
	if normalizedEmail == "" || strings.TrimSpace(password) == "" {
		return userdomain.User{}, userdomain.ErrInvalidInput
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return userdomain.User{}, err
	}

	active := true
	if isActive != nil {
		active = *isActive
	}

	normalizedRoles := normalizeIDs(roleIDs)
	return s.repo.CreateUser(ctx, normalizedEmail, passwordHash, active, normalizedRoles)
}

func (s *Service) UpdateUser(ctx context.Context, id string, email, password *string, isActive *bool) (userdomain.User, error) {
	if strings.TrimSpace(id) == "" {
		return userdomain.User{}, userdomain.ErrInvalidInput
	}
	if email == nil && password == nil && isActive == nil {
		return userdomain.User{}, userdomain.ErrInvalidInput
	}

	current, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return userdomain.User{}, err
	}

	bumpTokenVersion := false
	if email != nil {
		normalizedEmail := normalizeEmail(*email)
		if normalizedEmail == "" {
			return userdomain.User{}, userdomain.ErrInvalidInput
		}
		current.Email = normalizedEmail
	}

	if password != nil {
		if strings.TrimSpace(*password) == "" {
			return userdomain.User{}, userdomain.ErrInvalidInput
		}
		passwordHash, err := hashPassword(*password)
		if err != nil {
			return userdomain.User{}, err
		}
		current.PasswordHash = passwordHash
		bumpTokenVersion = true
	}

	if isActive != nil {
		if current.IsActive != *isActive {
			bumpTokenVersion = true
		}
		current.IsActive = *isActive
	}

	return s.repo.UpdateUser(ctx, id, current.Email, current.PasswordHash, current.IsActive, bumpTokenVersion)
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return userdomain.ErrInvalidInput
	}
	return s.repo.DeleteUser(ctx, id)
}

func (s *Service) ListUserRoles(ctx context.Context, userID string) ([]rbacdomain.Role, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, userdomain.ErrInvalidInput
	}
	return s.repo.ListUserRoles(ctx, userID)
}

func (s *Service) ReplaceUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	if strings.TrimSpace(userID) == "" {
		return userdomain.ErrInvalidInput
	}
	return s.repo.ReplaceUserRoles(ctx, userID, normalizeIDs(roleIDs))
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
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
