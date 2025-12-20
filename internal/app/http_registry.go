package app

import (
	"context"
	"errors"

	jwtinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/jwt"
	pgdb "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/postgres"
	redisinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/redis"
	xenditinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/xendit"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	postgresrepo "github.com/mzulfanw/boilerplate-go-fiber/internal/repository/postgres"
	authservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/auth"
	emailservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/email"
	paymentservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/payment"
	rbacservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/rbac"
	userservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/user"
	httptransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http"
	authtransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/auth"
	docstransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/docs"
	healthtransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/health"
	paymenttransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/payment"
	rbactransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/rbac"
	usertransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/user"
)

type httpRegistry struct {
	Routers     []httptransport.Router
	AuthService *authservice.Service
}

func httpRouters(cfg config.Config, db pgdb.DB, xenditClient xenditinfra.XenditClient, cache redisinfra.Cache, emailService *emailservice.Service, emailRenderer *emailservice.Renderer) (httpRegistry, error) {
	if db == nil || db.Pool() == nil {
		return httpRegistry{}, errors.New("postgres pool is nil")
	}
	if xenditClient == nil {
		return httpRegistry{}, errors.New("xendit client is nil")
	}

	authRepo := postgresrepo.NewAuthRepository(db.Pool())
	tokenManager, err := jwtinfra.NewManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTokenTTL)
	if err != nil {
		return httpRegistry{}, err
	}
	authService, err := authservice.NewServiceWithCache(cfg, authRepo, tokenManager, cache)
	if err != nil {
		return httpRegistry{}, err
	}
	authHandler := authtransport.NewHandler(authService, emailService, emailRenderer, cfg.AuthPasswordResetURL, cfg.AppName)
	authMiddleware := httptransport.NewAuthMiddleware(authService)

	rbacRepo := postgresrepo.NewRBACRepository(db.Pool())
	rbacService, err := rbacservice.NewService(rbacRepo)
	if err != nil {
		return httpRegistry{}, err
	}
	rbacHandler := rbactransport.NewHandler(rbacService)

	userRepo := postgresrepo.NewUserRepository(db.Pool())
	userService, err := userservice.NewService(userRepo)
	if err != nil {
		return httpRegistry{}, err
	}
	userHandler := usertransport.NewHandler(userService)

	paymentService, err := paymentservice.NewService(xenditClient, cache)
	if err != nil {
		return httpRegistry{}, err
	}
	paymentHandler := paymenttransport.NewHandler(paymentService, cfg.XENDIT_WEBHOOK_TOKEN)

	healthDependencies := []healthtransport.Dependency{
		{
			Name: "postgres",
			Check: func(ctx context.Context) error {
				if db == nil || db.Pool() == nil {
					return errors.New("postgres: pool is nil")
				}
				return db.Pool().Ping(ctx)
			},
		},
		{
			Name: "redis",
			Check: func(ctx context.Context) error {
				if cache == nil {
					return errors.New("redis: cache is nil")
				}
				return cache.Ping(ctx)
			},
		},
	}

	routers := []httptransport.Router{
		healthtransport.NewRouter(cfg, healthDependencies...),
		docstransport.NewRouter(cfg),
		authtransport.NewRouter(authHandler, cfg),
		rbactransport.NewRouter(rbacHandler, authMiddleware),
		usertransport.NewRouter(userHandler, authMiddleware),
		paymenttransport.NewRouter(paymentHandler),
	}

	return httpRegistry{
		Routers:     routers,
		AuthService: authService,
	}, nil
}
