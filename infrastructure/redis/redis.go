package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrKeyNotFound = errors.New("redis: key not found")
	ErrNilClient   = errors.New("redis: client is nil")
	ErrEmptyKey    = errors.New("redis: key is empty")
)

type Client struct {
	client     *goredis.Client
	defaultTTL time.Duration
}

var _ Cache = (*Client)(nil)

func New(cfg config.Config) (*Client, error) {
	if cfg.RedisAddr == "" {
		return nil, errors.New("redis: REDIS_ADDR is empty")
	}

	opts := &goredis.Options{
		Addr:         cfg.RedisAddr,
		Username:     cfg.RedisUsername,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
	}
	if cfg.RedisPoolSize > 0 {
		opts.PoolSize = cfg.RedisPoolSize
	}
	if cfg.RedisMinIdleConns > 0 {
		opts.MinIdleConns = cfg.RedisMinIdleConns
	}
	if cfg.RedisTLSEnabled {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: cfg.RedisTLSInsecureSkipVerify,
		}
	}

	client := goredis.NewClient(opts)
	pingTimeout := cfg.RedisDialTimeout
	if pingTimeout <= 0 {
		pingTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{
		client:     client,
		defaultTTL: cfg.RedisDefaultTTL,
	}, nil
}

func (c *Client) Set(key string, value interface{}) error {
	return c.SetWithTTL(key, value, c.defaultTTL)
}

func (c *Client) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	if err := c.validateKey(key); err != nil {
		return err
	}

	payload, err := marshalValue(value)
	if err != nil {
		return err
	}

	return c.client.Set(context.Background(), key, payload, ttl).Err()
}

func (c *Client) SetIfNotExists(key string, value interface{}, ttl time.Duration) (bool, error) {
	if err := c.validateKey(key); err != nil {
		return false, err
	}

	payload, err := marshalValue(value)
	if err != nil {
		return false, err
	}

	return c.client.SetNX(context.Background(), key, payload, ttl).Result()
}

func (c *Client) Get(key string) (interface{}, error) {
	return c.GetBytes(key)
}

func (c *Client) GetBytes(key string) ([]byte, error) {
	if err := c.validateKey(key); err != nil {
		return nil, err
	}

	data, err := c.client.Get(context.Background(), key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	return data, nil
}

func (c *Client) GetString(key string) (string, error) {
	data, err := c.GetBytes(key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return ErrNilClient
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return c.client.Ping(ctx).Err()
}

func (c *Client) LPush(key string, values ...interface{}) (int64, error) {
	if err := c.validateKey(key); err != nil {
		return 0, err
	}
	return c.client.LPush(context.Background(), key, values...).Result()
}

func (c *Client) LRange(key string, start, stop int64) ([]string, error) {
	if err := c.validateKey(key); err != nil {
		return nil, err
	}
	return c.client.LRange(context.Background(), key, start, stop).Result()
}

func (c *Client) LRem(key string, count int64, value interface{}) (int64, error) {
	if err := c.validateKey(key); err != nil {
		return 0, err
	}
	return c.client.LRem(context.Background(), key, count, value).Result()
}

func (c *Client) BRPopLPush(source, destination string, timeout time.Duration) (string, error) {
	if err := c.validateKey(source); err != nil {
		return "", err
	}
	if err := c.validateKey(destination); err != nil {
		return "", err
	}

	data, err := c.client.BRPopLPush(context.Background(), source, destination, timeout).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ErrKeyNotFound
		}
		return "", err
	}
	return data, nil
}

func (c *Client) ZAdd(key string, members ...ZMember) (int64, error) {
	if err := c.validateKey(key); err != nil {
		return 0, err
	}
	if len(members) == 0 {
		return 0, nil
	}

	items := make([]goredis.Z, 0, len(members))
	for _, member := range members {
		items = append(items, goredis.Z{
			Score:  member.Score,
			Member: member.Member,
		})
	}

	return c.client.ZAdd(context.Background(), key, items...).Result()
}

func (c *Client) ZRangeByScore(key, min, max string, count int64) ([]string, error) {
	if err := c.validateKey(key); err != nil {
		return nil, err
	}

	args := &goredis.ZRangeBy{
		Min: min,
		Max: max,
	}
	if count > 0 {
		args.Count = count
	}

	return c.client.ZRangeByScore(context.Background(), key, args).Result()
}

func (c *Client) ZRem(key string, members ...interface{}) (int64, error) {
	if err := c.validateKey(key); err != nil {
		return 0, err
	}
	if len(members) == 0 {
		return 0, nil
	}
	return c.client.ZRem(context.Background(), key, members...).Result()
}

func (c *Client) GetJSON(key string, dest interface{}) error {
	data, err := c.GetBytes(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *Client) Delete(key string) error {
	if err := c.validateKey(key); err != nil {
		return err
	}
	return c.client.Del(context.Background(), key).Err()
}

func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *Client) validateKey(key string) error {
	if c == nil || c.client == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	return nil
}

func marshalValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case nil:
		return nil, errors.New("redis: value is nil")
	case []byte:
		return v, nil
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("redis: marshal value: %w", err)
		}
		return raw, nil
	}
}
