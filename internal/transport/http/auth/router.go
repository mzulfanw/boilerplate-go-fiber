package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
)

type Router struct {
	handler      *Handler
	loginLimiter fiber.Handler
}

func NewRouter(handler *Handler, cfg config.Config) *Router {
	var loginLimiter fiber.Handler
	if cfg.AuthLoginRateLimit > 0 && cfg.AuthLoginRateWindow > 0 {
		loginLimiter = limiter.New(limiter.Config{
			Max:        cfg.AuthLoginRateLimit,
			Expiration: cfg.AuthLoginRateWindow,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				return fiber.NewError(fiber.StatusTooManyRequests, "too many login attempts")
			},
		})
	}

	return &Router{
		handler:      handler,
		loginLimiter: loginLimiter,
	}
}

func (r *Router) Register(app *fiber.App) {
	if r == nil || r.handler == nil || app == nil {
		return
	}

	group := app.Group("/auth")
	if r.loginLimiter != nil {
		group.Post("/login", r.loginLimiter, r.handler.Login)
	} else {
		group.Post("/login", r.handler.Login)
	}
	group.Post("/refresh", r.handler.Refresh)
	group.Post("/logout", r.handler.Logout)
	group.Post("/forgot-password", r.handler.ForgotPassword)
	group.Post("/reset-password", r.handler.ResetPassword)
}
