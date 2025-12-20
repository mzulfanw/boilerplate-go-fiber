package http

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

const requestIDKey = "requestid"

func setupMiddlewares(app *fiber.App, cfg config.Config) {
	if cfg.MetricsEnabled && strings.TrimSpace(cfg.MetricsPath) != "" {
		app.Get(cfg.MetricsPath, monitor.New(monitor.Config{
			Title: "Metrics",
		}))
	}

	if cfg.TracingEnabled {
		app.Use(otelfiber.Middleware(
			otelfiber.WithTracerProvider(otel.GetTracerProvider()),
		))
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     cfg.CORSAllowMethods,
		AllowHeaders:     cfg.CORSAllowHeaders,
		ExposeHeaders:    cfg.CORSExposeHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           cfg.CORSMaxAge,
	}))
	app.Use(helmet.New())
	app.Use(requestid.New())
	app.Use(requestLogger())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			logrus.WithFields(logrus.Fields{
				"panic":      fmt.Sprint(e),
				"request_id": getRequestID(c),
			}).Error(string(debug.Stack()))
		},
	}))
}

func requestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		status := c.Response().StatusCode()
		entry := logrus.WithFields(logrus.Fields{
			"status":      status,
			"method":      c.Method(),
			"path":        c.Path(),
			"latency_ms":  float64(latency.Microseconds()) / 1000.0,
			"ip":          c.IP(),
			"user_agent":  c.Get(fiber.HeaderUserAgent),
			"request_id":  getRequestID(c),
			"status_text": utils.StatusMessage(status),
		})
		if err != nil {
			entry = entry.WithError(err)
		}

		switch {
		case status >= fiber.StatusInternalServerError:
			entry.Error("request completed")
		case status >= fiber.StatusBadRequest:
			entry.Warn("request completed")
		default:
			entry.Info("request completed")
		}

		return err
	}
}

func getRequestID(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}
	if value := c.Locals(requestIDKey); value != nil {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return c.Get(fiber.HeaderXRequestID)
}
