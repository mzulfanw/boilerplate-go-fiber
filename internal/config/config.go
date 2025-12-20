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

	PostgresDSN               string
	PostgresHost              string
	PostgresPort              int
	PostgresUser              string
	PostgresPassword          string
	PostgresDB                string
	PostgresSSLMode           string
	PostgresConnectTimeout    time.Duration
	PostgresMaxConns          int
	PostgresMinConns          int
	PostgresMaxConnLifetime   time.Duration
	PostgresMaxConnIdleTime   time.Duration
	PostgresHealthCheckPeriod time.Duration

	JWTSecret       string
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	RefreshTokenCleanupInterval time.Duration

	AuthMaxLoginAttempts int
	AuthLockoutDuration  time.Duration
	AuthLoginRateLimit   int
	AuthLoginRateWindow  time.Duration

	MetricsEnabled bool
	MetricsPath    string

	SwaggerEnabled bool
	SwaggerPath    string

	TracingEnabled       bool
	OTELExporterEndpoint string
	OTELExporterInsecure bool
	OTELSampleRatio      float64
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

		PostgresDSN:      getString("POSTGRES_DSN", ""),
		PostgresHost:     getString("POSTGRES_HOST", "localhost"),
		PostgresUser:     getString("POSTGRES_USER", "postgres"),
		PostgresPassword: getString("POSTGRES_PASSWORD", ""),
		PostgresDB:       getString("POSTGRES_DB", "postgres"),
		PostgresSSLMode:  getString("POSTGRES_SSLMODE", "disable"),

		JWTSecret: getString("JWT_SECRET", ""),
		JWTIssuer: getString("JWT_ISSUER", "boilerplate-go-fiber"),

		MetricsPath:          getString("METRICS_PATH", "/metrics"),
		SwaggerPath:          getString("SWAGGER_PATH", "/docs"),
		OTELExporterEndpoint: getString("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
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
	if cfg.PostgresPort, err = getInt("POSTGRES_PORT", 5432); err != nil {
		return Config{}, err
	}
	if cfg.PostgresMaxConns, err = getInt("POSTGRES_MAX_CONNS", 10); err != nil {
		return Config{}, err
	}
	if cfg.PostgresMinConns, err = getInt("POSTGRES_MIN_CONNS", 0); err != nil {
		return Config{}, err
	}
	if cfg.PostgresConnectTimeout, err = getDuration("POSTGRES_CONNECT_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.PostgresMaxConnLifetime, err = getDuration("POSTGRES_MAX_CONN_LIFETIME", 0); err != nil {
		return Config{}, err
	}
	if cfg.PostgresMaxConnIdleTime, err = getDuration("POSTGRES_MAX_CONN_IDLE_TIME", 0); err != nil {
		return Config{}, err
	}
	if cfg.PostgresHealthCheckPeriod, err = getDuration("POSTGRES_HEALTHCHECK_PERIOD", time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.AccessTokenTTL, err = getDuration("ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.RefreshTokenTTL, err = getDuration("REFRESH_TOKEN_TTL", 168*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.RefreshTokenCleanupInterval, err = getDuration("REFRESH_TOKEN_CLEANUP_INTERVAL", time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.AuthLockoutDuration, err = getDuration("AUTH_LOCKOUT_DURATION", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.AuthLoginRateWindow, err = getDuration("AUTH_LOGIN_RATE_WINDOW", time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.AuthMaxLoginAttempts, err = getInt("AUTH_MAX_LOGIN_ATTEMPTS", 5); err != nil {
		return Config{}, err
	}
	if cfg.AuthLoginRateLimit, err = getInt("AUTH_LOGIN_RATE_LIMIT", 10); err != nil {
		return Config{}, err
	}
	if cfg.MetricsEnabled, err = getBool("METRICS_ENABLED", true); err != nil {
		return Config{}, err
	}
	if cfg.SwaggerEnabled, err = getBool("SWAGGER_ENABLED", false); err != nil {
		return Config{}, err
	}
	if cfg.TracingEnabled, err = getBool("OTEL_ENABLED", false); err != nil {
		return Config{}, err
	}
	if cfg.OTELExporterInsecure, err = getBool("OTEL_EXPORTER_OTLP_INSECURE", true); err != nil {
		return Config{}, err
	}
	if cfg.OTELSampleRatio, err = getFloat("OTEL_SAMPLE_RATIO", 1.0); err != nil {
		return Config{}, err
	}

	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	if cfg.CORSAllowCredentials && strings.TrimSpace(cfg.CORSAllowOrigins) == "*" {
		return Config{}, fmt.Errorf("CORS_ALLOW_CREDENTIALS=true requires CORS_ALLOW_ORIGINS not '*'")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.OTELSampleRatio < 0 || cfg.OTELSampleRatio > 1 {
		return Config{}, fmt.Errorf("OTEL_SAMPLE_RATIO must be between 0 and 1")
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

func getFloat(key string, fallback float64) (float64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
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
