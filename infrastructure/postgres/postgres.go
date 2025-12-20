package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
)

type Client struct {
	pool *pgxpool.Pool
}

var _ DB = (*Client)(nil)

func New(cfg config.Config) (*Client, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}

	pgConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if cfg.PostgresConnectTimeout > 0 {
		pgConfig.ConnConfig.ConnectTimeout = cfg.PostgresConnectTimeout
	}
	if cfg.PostgresMaxConns > 0 {
		pgConfig.MaxConns = int32(cfg.PostgresMaxConns)
	}
	if cfg.PostgresMinConns > 0 {
		pgConfig.MinConns = int32(cfg.PostgresMinConns)
	}
	if cfg.PostgresMaxConnLifetime > 0 {
		pgConfig.MaxConnLifetime = cfg.PostgresMaxConnLifetime
	}
	if cfg.PostgresMaxConnIdleTime > 0 {
		pgConfig.MaxConnIdleTime = cfg.PostgresMaxConnIdleTime
	}
	if cfg.PostgresHealthCheckPeriod > 0 {
		pgConfig.HealthCheckPeriod = cfg.PostgresHealthCheckPeriod
	}

	pingTimeout := cfg.PostgresConnectTimeout
	if pingTimeout <= 0 {
		pingTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &Client{pool: pool}, nil
}

func (c *Client) Pool() *pgxpool.Pool {
	if c == nil {
		return nil
	}
	return c.pool
}

func (c *Client) Close() {
	if c == nil || c.pool == nil {
		return
	}
	c.pool.Close()
}

func buildDSN(cfg config.Config) (string, error) {
	if cfg.PostgresDSN != "" {
		return cfg.PostgresDSN, nil
	}

	if cfg.PostgresHost == "" {
		return "", errors.New("postgres: POSTGRES_HOST is empty")
	}
	if cfg.PostgresUser == "" {
		return "", errors.New("postgres: POSTGRES_USER is empty")
	}
	if cfg.PostgresDB == "" {
		return "", errors.New("postgres: POSTGRES_DB is empty")
	}

	port := cfg.PostgresPort
	if port <= 0 {
		port = 5432
	}

	sslMode := cfg.PostgresSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	userInfo := url.User(cfg.PostgresUser)
	if cfg.PostgresPassword != "" {
		userInfo = url.UserPassword(cfg.PostgresUser, cfg.PostgresPassword)
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   userInfo,
		Host:   fmt.Sprintf("%s:%d", cfg.PostgresHost, port),
		Path:   cfg.PostgresDB,
	}

	query := url.Values{}
	query.Set("sslmode", sslMode)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
