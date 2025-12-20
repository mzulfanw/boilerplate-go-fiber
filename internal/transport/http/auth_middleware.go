package http

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	authusecase "github.com/mzulfanw/boilerplate-go-fiber/internal/service/auth"
)

const authContextKey = "auth_context"

type AuthContext struct {
	UserID      string
	Roles       []string
	Permissions []string
}

type AuthMiddleware struct {
	service *authusecase.Service
}

func NewAuthMiddleware(service *authusecase.Service) *AuthMiddleware {
	return &AuthMiddleware{service: service}
}

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if m == nil || m.service == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "auth middleware not configured")
		}
		if _, err := m.ensureAuthContext(c); err != nil {
			return err
		}
		return c.Next()
	}
}

func (m *AuthMiddleware) RequirePermissions(permissions ...string) fiber.Handler {
	required := normalizePermissions(permissions)
	return func(c *fiber.Ctx) error {
		if m == nil || m.service == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "auth middleware not configured")
		}
		ctx, err := m.ensureAuthContext(c)
		if err != nil {
			return err
		}
		if len(required) == 0 {
			return c.Next()
		}
		if !hasAnyPermission(ctx.Permissions, required) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}
		return c.Next()
	}
}

func GetAuthContext(c *fiber.Ctx) (AuthContext, bool) {
	if c == nil {
		return AuthContext{}, false
	}
	value := c.Locals(authContextKey)
	if value == nil {
		return AuthContext{}, false
	}
	ctx, ok := value.(AuthContext)
	return ctx, ok
}

func (m *AuthMiddleware) ensureAuthContext(c *fiber.Ctx) (AuthContext, error) {
	if ctx, ok := GetAuthContext(c); ok {
		return ctx, nil
	}
	return m.authenticate(c)
}

func (m *AuthMiddleware) authenticate(c *fiber.Ctx) (AuthContext, error) {
	if c == nil {
		return AuthContext{}, fiber.NewError(fiber.StatusUnauthorized, "invalid request")
	}
	authHeader := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
	if authHeader == "" {
		return AuthContext{}, fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
	}
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return AuthContext{}, fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header")
	}

	claims, err := m.service.ValidateAccessToken(c.UserContext(), parts[1])
	if err != nil {
		if errors.Is(err, authusecase.ErrUserDisabled) {
			return AuthContext{}, fiber.NewError(fiber.StatusForbidden, "user is disabled")
		}
		return AuthContext{}, fiber.NewError(fiber.StatusUnauthorized, "invalid or expired access token")
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return AuthContext{}, fiber.NewError(fiber.StatusUnauthorized, "invalid access token")
	}

	ctx := AuthContext{
		UserID:      claims.Subject,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}
	c.Locals(authContextKey, ctx)
	return ctx, nil
}

func normalizePermissions(permissions []string) []string {
	if len(permissions) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(permissions))
	for _, perm := range permissions {
		normalized := strings.ToLower(strings.TrimSpace(perm))
		if normalized == "" {
			continue
		}
		seen[normalized] = struct{}{}
	}
	if len(seen) == 0 {
		return nil
	}
	result := make([]string, 0, len(seen))
	for perm := range seen {
		result = append(result, perm)
	}
	return result
}

func hasAnyPermission(granted, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(granted) == 0 {
		return false
	}
	allowed := make(map[string]struct{}, len(granted))
	for _, perm := range granted {
		normalized := strings.ToLower(strings.TrimSpace(perm))
		if normalized == "" {
			continue
		}
		allowed[normalized] = struct{}{}
	}
	for _, perm := range required {
		if _, ok := allowed[strings.ToLower(strings.TrimSpace(perm))]; ok {
			return true
		}
	}
	return false
}
