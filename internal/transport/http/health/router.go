package health

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
)

type Router struct {
	appName   string
	env       string
	startedAt time.Time
	checks    []Dependency
	timeout   time.Duration
}

type Dependency struct {
	Name  string
	Check func(ctx context.Context) error
}

const defaultCheckTimeout = 2 * time.Second

func NewRouter(cfg config.Config, dependencies ...Dependency) *Router {
	return &Router{
		appName:   cfg.AppName,
		env:       cfg.Env,
		startedAt: time.Now(),
		checks:    dependencies,
		timeout:   defaultCheckTimeout,
	}
}

func (r *Router) Register(app *fiber.App) {
	if r == nil || app == nil {
		return
	}

	app.Get("/health", r.Liveness)
	app.Get("/health/liveness", r.Liveness)
	app.Get("/health/readiness", r.Readiness)
}

// Liveness godoc
// @Summary Liveness check
// @Tags Health
// @Produce json
// @Success 200 {object} response.Response{data=response.Status}
// @Router /health [get]
// @Router /health/liveness [get]
func (r *Router) Liveness(c *fiber.Ctx) error {
	now := time.Now().UTC()
	status := response.Status{
		Status: "ok",
		App:    r.appName,
		Env:    r.env,
		Uptime: time.Since(r.startedAt).Round(time.Second).String(),
		Time:   now.Format(time.RFC3339),
	}
	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    status,
	}
	return c.Status(resp.Code).JSON(resp)
}

// Readiness godoc
// @Summary Readiness check
// @Tags Health
// @Produce json
// @Success 200 {object} response.Response{data=response.ReadinessStatus}
// @Failure 503 {object} response.Response
// @Router /health/readiness [get]
func (r *Router) Readiness(c *fiber.Ctx) error {
	now := time.Now().UTC()
	uptime := time.Since(r.startedAt).Round(time.Second).String()
	checks := make([]response.DependencyStatus, 0, len(r.checks))
	ready := true

	for _, dep := range r.checks {
		name := strings.TrimSpace(dep.Name)
		if name == "" {
			name = "dependency"
		}
		status := response.DependencyStatus{
			Name: name,
		}

		if dep.Check == nil {
			status.Status = "error"
			status.Error = "no check configured"
			ready = false
			checks = append(checks, status)
			continue
		}

		ctx := c.UserContext()
		if r.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, r.timeout)
			start := time.Now()
			err := dep.Check(ctx)
			cancel()
			status.Latency = time.Since(start).Round(time.Millisecond).String()
			if err != nil {
				status.Status = "error"
				status.Error = err.Error()
				ready = false
			} else {
				status.Status = "ok"
			}
		} else {
			start := time.Now()
			err := dep.Check(ctx)
			status.Latency = time.Since(start).Round(time.Millisecond).String()
			if err != nil {
				status.Status = "error"
				status.Error = err.Error()
				ready = false
			} else {
				status.Status = "ok"
			}
		}

		checks = append(checks, status)
	}

	overallStatus := "ok"
	code := fiber.StatusOK
	message := "ok"
	if !ready {
		overallStatus = "fail"
		code = fiber.StatusServiceUnavailable
		message = "not ready"
	}

	payload := response.ReadinessStatus{
		Status:       overallStatus,
		App:          r.appName,
		Env:          r.env,
		Uptime:       uptime,
		Time:         now.Format(time.RFC3339),
		Dependencies: checks,
	}

	resp := response.Response{
		Code:    code,
		Message: message,
		Data:    payload,
	}
	return c.Status(code).JSON(resp)
}
