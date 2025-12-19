package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	"github.com/sirupsen/logrus"
)

type Server struct {
	app  *fiber.App
	addr string
}

func NewServer(cfg config.Config, routers ...Router) *Server {
	app := fiber.New(fiber.Config{
		IdleTimeout:  cfg.HTTPIdleTimeout,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		ErrorHandler: errorHandler(),
	})

	setupMiddlewares(app, cfg)
	for _, r := range routers {
		if r == nil {
			continue
		}
		r.Register(app)
	}

	return &Server{
		app:  app,
		addr: cfg.HTTPAddr,
	}
}

func (s *Server) Start() error {
	logrus.WithField("addr", s.addr).Info("fiber running")
	return s.app.Listen(s.addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
