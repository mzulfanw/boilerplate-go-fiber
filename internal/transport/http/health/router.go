package health

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
)

type Router struct {
	appName   string
	env       string
	startedAt time.Time
}

func NewRouter(cfg config.Config) *Router {
	return &Router{
		appName:   cfg.AppName,
		env:       cfg.Env,
		startedAt: time.Now(),
	}
}

func (r *Router) Register(app *fiber.App) {
	if r == nil || app == nil {
		return
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		status := response.Status{
			Status: "ok",
			App:    r.appName,
			Env:    r.env,
			Uptime: time.Since(r.startedAt).Round(time.Second).String(),
			Time:   time.Now().UTC().Format(time.RFC3339),
		}
		resp := response.Response{
			Code:    fiber.StatusOK,
			Message: "ok",
			Data:    status,
		}
		return c.Status(resp.Code).JSON(resp)
	})
}
