package redis

import (
	"context"
	"time"
)

type Cache interface {
	Set(key string, value interface{}) error
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	SetIfNotExists(key string, value interface{}, ttl time.Duration) (bool, error)
	Get(key string) (interface{}, error)
	GetString(key string) (string, error)
	Delete(key string) error
	Ping(ctx context.Context) error
	LPush(key string, values ...interface{}) (int64, error)
	LRange(key string, start, stop int64) ([]string, error)
	LRem(key string, count int64, value interface{}) (int64, error)
	BRPopLPush(source, destination string, timeout time.Duration) (string, error)
	ZAdd(key string, members ...ZMember) (int64, error)
	ZRangeByScore(key, min, max string, count int64) ([]string, error)
	ZRem(key string, members ...interface{}) (int64, error)
	Close() error
}

type ZMember struct {
	Score  float64
	Member interface{}
}
