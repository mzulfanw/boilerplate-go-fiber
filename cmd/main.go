package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mzulfanw/boilerplate-go-fiber/internal/app"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/logger"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/observability"
	"github.com/sirupsen/logrus"
)

// @title Boilerplate Go Fiber API
// @version 0.1.0
// @description API boilerplate with Go + Fiber.
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("failed to load config")
	}

	log := logger.Init(cfg)

	shutdownTracing, err := observability.InitTracing(context.Background(), cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize tracing")
	}
	if shutdownTracing != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdownTracing(ctx); err != nil {
				log.WithError(err).Warn("failed to shutdown tracing")
			}
		}()
	}

	application, err := app.NewApp(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize app")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		sig := <-sigChan
		log.WithField("signal", sig.String()).Info("received signal")
		cancel()
	}()

	if err := application.Run(ctx); err != nil {
		log.WithError(err).Error("application error")
		return
	}

	log.Info("shutdown complete")
}
