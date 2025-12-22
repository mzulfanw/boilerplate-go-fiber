package redis

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	SetIfNotExists(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	Get(ctx context.Context, key string) (interface{}, error)
	GetString(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LRem(ctx context.Context, key string, count int64, value interface{}) (int64, error)
	BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) (string, error)
	ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error)
	ZRangeByScore(ctx context.Context, key, min, max string, count int64) ([]string, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	Close() error
}

type ZMember struct {
	Score  float64
	Member interface{}
}
