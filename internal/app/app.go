package app

import (
	"context"
	"time"

	"github.com/mzulfanw/boilerplate-go-fiber/infrastructure/redis"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	httptransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http"
	"github.com/sirupsen/logrus"
)

type App struct {
	httpServer      *httptransport.Server
	shutdownTimeout time.Duration
	cache           redis.Cache
}

func NewApp(cfg config.Config) (*App, error) {
	routers := httpRouters(cfg)
	httpServer := httptransport.NewServer(cfg, routers...)
	cache, err := redis.New(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		httpServer:      httpServer,
		shutdownTimeout: cfg.ShutdownTimeout,
		cache:           cache,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.closeResources()

	errChan := make(chan error, 1)

	go func() {
		errChan <- a.httpServer.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
		defer cancel()
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errChan
	case err := <-errChan:
		return err
	}
}

func (a *App) closeResources() {
	if a.cache == nil {
		return
	}
	if err := a.cache.Close(); err != nil {
		logrus.WithError(err).Warn("redis close failed")
	}
}
