package app

import (
	"errors"

	jwtinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/jwt"
	pgdb "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/postgres"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	postgresrepo "github.com/mzulfanw/boilerplate-go-fiber/internal/repository/postgres"
	authservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/auth"
	rbacservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/rbac"
	userservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/user"
	httptransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http"
	authtransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/auth"
	docstransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/docs"
	healthtransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/health"
	rbactransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/rbac"
	usertransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/user"
)

type httpRegistry struct {
	Routers     []httptransport.Router
	AuthService *authservice.Service
}

func httpRouters(cfg config.Config, db pgdb.DB) (httpRegistry, error) {
	if db == nil || db.Pool() == nil {
		return httpRegistry{}, errors.New("postgres pool is nil")
	}

	authRepo := postgresrepo.NewAuthRepository(db.Pool())
	tokenManager, err := jwtinfra.NewManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTokenTTL)
	if err != nil {
		return httpRegistry{}, err
	}
	authService, err := authservice.NewService(cfg, authRepo, tokenManager)
	if err != nil {
		return httpRegistry{}, err
	}
	authHandler := authtransport.NewHandler(authService)
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

	routers := []httptransport.Router{
		healthtransport.NewRouter(cfg),
		docstransport.NewRouter(cfg),
		authtransport.NewRouter(authHandler, cfg),
		rbactransport.NewRouter(rbacHandler, authMiddleware),
		usertransport.NewRouter(userHandler, authMiddleware),
	}

	return httpRegistry{
		Routers:     routers,
		AuthService: authService,
	}, nil
}
