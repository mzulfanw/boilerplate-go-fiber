package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	Env     string

	LogLevel string

	HTTPAddr         string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration

	ShutdownTimeout time.Duration

	CORSAllowOrigins     string
	CORSAllowMethods     string
	CORSAllowHeaders     string
	CORSExposeHeaders    string
	CORSAllowCredentials bool
	CORSMaxAge           int

	RedisAddr                  string
	RedisUsername              string
	RedisPassword              string
	RedisDB                    int
	RedisDialTimeout           time.Duration
	RedisReadTimeout           time.Duration
	RedisWriteTimeout          time.Duration
	RedisPoolSize              int
	RedisMinIdleConns          int
	RedisDefaultTTL            time.Duration
	RedisTLSEnabled            bool
	RedisTLSInsecureSkipVerify bool
}

func Load() (Config, error) {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = "local"
	}

	if env == "local" {
		if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		if loadedEnv := strings.TrimSpace(os.Getenv("APP_ENV")); loadedEnv != "" {
			env = loadedEnv
		}
	}

	cfg := Config{
		AppName: getString("APP_NAME", "boilerplate-go-fiber"),
		Env:     env,

		LogLevel: getString("LOG_LEVEL", "info"),

		HTTPAddr: getString("HTTP_ADDR", ":3000"),

		CORSAllowOrigins:  getString("CORS_ALLOW_ORIGINS", "*"),
		CORSAllowMethods:  getString("CORS_ALLOW_METHODS", "GET,POST,HEAD,PUT,DELETE,PATCH"),
		CORSAllowHeaders:  getString("CORS_ALLOW_HEADERS", ""),
		CORSExposeHeaders: getString("CORS_EXPOSE_HEADERS", ""),

		RedisAddr:     getString("REDIS_ADDR", "localhost:6379"),
		RedisUsername: getString("REDIS_USERNAME", ""),
		RedisPassword: getString("REDIS_PASSWORD", ""),
	}

	var err error
	if cfg.HTTPReadTimeout, err = getDuration("HTTP_READ_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTPWriteTimeout, err = getDuration("HTTP_WRITE_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTPIdleTimeout, err = getDuration("HTTP_IDLE_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = getDuration("SHUTDOWN_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.CORSAllowCredentials, err = getBool("CORS_ALLOW_CREDENTIALS", false); err != nil {
		return Config{}, err
	}
	if cfg.CORSMaxAge, err = getInt("CORS_MAX_AGE", 0); err != nil {
		return Config{}, err
	}
	if cfg.RedisDB, err = getInt("REDIS_DB", 0); err != nil {
		return Config{}, err
	}
	if cfg.RedisPoolSize, err = getInt("REDIS_POOL_SIZE", 0); err != nil {
		return Config{}, err
	}
	if cfg.RedisMinIdleConns, err = getInt("REDIS_MIN_IDLE_CONNS", 0); err != nil {
		return Config{}, err
	}
	if cfg.RedisDialTimeout, err = getDuration("REDIS_DIAL_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.RedisReadTimeout, err = getDuration("REDIS_READ_TIMEOUT", 3*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.RedisWriteTimeout, err = getDuration("REDIS_WRITE_TIMEOUT", 3*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.RedisDefaultTTL, err = getDuration("REDIS_DEFAULT_TTL", 0); err != nil {
		return Config{}, err
	}
	if cfg.RedisTLSEnabled, err = getBool("REDIS_TLS_ENABLED", false); err != nil {
		return Config{}, err
	}
	if cfg.RedisTLSInsecureSkipVerify, err = getBool("REDIS_TLS_INSECURE_SKIP_VERIFY", false); err != nil {
		return Config{}, err
	}

	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	if cfg.CORSAllowCredentials && strings.TrimSpace(cfg.CORSAllowOrigins) == "*" {
		return Config{}, fmt.Errorf("CORS_ALLOW_CREDENTIALS=true requires CORS_ALLOW_ORIGINS not '*'")
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return parsed, nil
}

func getBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s: %w", key, err)
	}

	return parsed, nil
}

func getInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return parsed, nil
}
