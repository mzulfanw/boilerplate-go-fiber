package app

import (
	"context"
	"time"

	emailinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/email"
	"github.com/mzulfanw/boilerplate-go-fiber/infrastructure/postgres"
	"github.com/mzulfanw/boilerplate-go-fiber/infrastructure/redis"
	"github.com/mzulfanw/boilerplate-go-fiber/infrastructure/xendit"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	authservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/auth"
	emailservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/email"
	httptransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http"
	"github.com/sirupsen/logrus"
)

type App struct {
	httpServer      *httptransport.Server
	shutdownTimeout time.Duration
	cache           redis.Cache
	db              postgres.DB
	xenditClient    xendit.XenditClient
	authService     *authservice.Service
	emailWorker     *emailservice.Worker

	refreshTokenCleanupInterval time.Duration
}

func NewApp(cfg config.Config) (*App, error) {
	cache, err := redis.New(cfg)
	if err != nil {
		return nil, err
	}
	db, err := postgres.New(cfg)
	if err != nil {
		_ = cache.Close()
		return nil, err
	}
	xenditClient, err := xendit.New(cfg)
	if err != nil {
		_ = cache.Close()
		db.Close()
		return nil, err
	}

	emailQueue := emailservice.NewRedisQueue(cache, emailservice.QueueOptions{
		Prefix: "email:queue",
	})
	emailSender, err := emailinfra.NewSMTP(cfg)
	if err != nil {
		_ = cache.Close()
		db.Close()
		return nil, err
	}

	var emailWorker *emailservice.Worker
	var emailService *emailservice.Service
	var emailRenderer *emailservice.Renderer
	if emailSender != nil {
		emailService = emailservice.NewService(emailQueue)
		emailRenderer = emailservice.NewRenderer(cfg.EmailTemplateDir)
		emailWorker = emailservice.NewWorker(emailQueue, emailSender, emailservice.WorkerOptions{
			RecoverInFlight: true,
		})
	}

	registry, err := httpRouters(cfg, db, xenditClient, cache, emailService, emailRenderer)
	if err != nil {
		_ = cache.Close()
		db.Close()
		return nil, err
	}
	httpServer := httptransport.NewServer(cfg, registry.Routers...)

	return &App{
		httpServer:                  httpServer,
		shutdownTimeout:             cfg.ShutdownTimeout,
		cache:                       cache,
		db:                          db,
		xenditClient:                xenditClient,
		authService:                 registry.AuthService,
		emailWorker:                 emailWorker,
		refreshTokenCleanupInterval: cfg.RefreshTokenCleanupInterval,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.closeResources()

	errChan := make(chan error, 1)

	a.startRefreshTokenCleanup(ctx)
	a.startEmailWorker(ctx)

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

func (a *App) startRefreshTokenCleanup(ctx context.Context) {
	if a == nil || a.authService == nil || a.refreshTokenCleanupInterval <= 0 {
		return
	}

	ticker := time.NewTicker(a.refreshTokenCleanupInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if ctx.Err() != nil {
					return
				}
				cleanupCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
				deleted, err := a.authService.CleanupExpiredRefreshTokens(cleanupCtx)
				cancel()
				if err != nil {
					logrus.WithError(err).Warn("refresh token cleanup failed")
					continue
				}
				if deleted > 0 {
					logrus.WithField("deleted", deleted).Info("refresh token cleanup completed")
				}
			}
		}
	}()
}

func (a *App) startEmailWorker(ctx context.Context) {
	if a == nil || a.emailWorker == nil {
		return
	}
	go a.emailWorker.Run(ctx)
}

func (a *App) closeResources() {
	if a.cache == nil {
	} else if err := a.cache.Close(); err != nil {
		logrus.WithError(err).Warn("redis close failed")
	}

	if a.db == nil {
		return
	}
	a.db.Close()
}
