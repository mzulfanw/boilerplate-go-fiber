package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	redisinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/redis"
)

var ErrQueueEmpty = errors.New("email queue is empty")
var ErrInvalidPayload = errors.New("email queue payload is invalid")

type Queue interface {
	Enqueue(ctx context.Context, job Job) error
	Reserve(ctx context.Context, timeout time.Duration) (Job, error)
	Ack(ctx context.Context, job Job) error
	Retry(ctx context.Context, job Job, delay time.Duration) error
	DeadLetter(ctx context.Context, job Job, reason string) error
	RequeueDue(ctx context.Context, limit int64) (int, error)
	RecoverInFlight(ctx context.Context) (int, error)
}

type QueueOptions struct {
	Prefix string
}

type RedisQueue struct {
	cache         redisinfra.Cache
	pendingKey    string
	processingKey string
	retryKey      string
	deadKey       string
}

func NewRedisQueue(cache redisinfra.Cache, opts QueueOptions) *RedisQueue {
	prefix := strings.TrimSpace(opts.Prefix)
	if prefix == "" {
		prefix = "email:queue"
	}

	return &RedisQueue{
		cache:         cache,
		pendingKey:    prefix + ":pending",
		processingKey: prefix + ":processing",
		retryKey:      prefix + ":retry",
		deadKey:       prefix + ":dead",
	}
}

func (q *RedisQueue) Enqueue(ctx context.Context, job Job) error {
	if q == nil || q.cache == nil {
		return errors.New("email queue: cache is nil")
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("email queue: marshal job: %w", err)
	}

	_, err = q.cache.LPush(q.pendingKey, string(payload))
	return err
}

func (q *RedisQueue) Reserve(ctx context.Context, timeout time.Duration) (Job, error) {
	if q == nil || q.cache == nil {
		return Job{}, errors.New("email queue: cache is nil")
	}

	raw, err := q.cache.BRPopLPush(q.pendingKey, q.processingKey, timeout)
	if err != nil {
		if errors.Is(err, redisinfra.ErrKeyNotFound) {
			return Job{}, ErrQueueEmpty
		}
		return Job{}, err
	}

	job := Job{Raw: raw}
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return job, ErrInvalidPayload
	}

	job.Raw = raw
	return job, nil
}

func (q *RedisQueue) Ack(ctx context.Context, job Job) error {
	if q == nil || q.cache == nil {
		return errors.New("email queue: cache is nil")
	}
	if strings.TrimSpace(job.Raw) == "" {
		return errors.New("email queue: empty job payload")
	}

	_, err := q.cache.LRem(q.processingKey, 1, job.Raw)
	return err
}

func (q *RedisQueue) Retry(ctx context.Context, job Job, delay time.Duration) error {
	if q == nil || q.cache == nil {
		return errors.New("email queue: cache is nil")
	}
	if strings.TrimSpace(job.Raw) == "" {
		return errors.New("email queue: empty job payload")
	}

	_, err := q.cache.LRem(q.processingKey, 1, job.Raw)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("email queue: marshal retry job: %w", err)
	}

	score := float64(time.Now().Add(delay).Unix())
	_, err = q.cache.ZAdd(q.retryKey, redisinfra.ZMember{
		Score:  score,
		Member: string(payload),
	})
	return err
}

func (q *RedisQueue) DeadLetter(ctx context.Context, job Job, reason string) error {
	if q == nil || q.cache == nil {
		return errors.New("email queue: cache is nil")
	}
	if strings.TrimSpace(job.Raw) == "" {
		return errors.New("email queue: empty job payload")
	}

	_, err := q.cache.LRem(q.processingKey, 1, job.Raw)
	if err != nil {
		return err
	}

	job.LastError = reason
	job.UpdatedAt = time.Now().UTC()
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("email queue: marshal dead-letter job: %w", err)
	}

	_, err = q.cache.LPush(q.deadKey, string(payload))
	return err
}

func (q *RedisQueue) RequeueDue(ctx context.Context, limit int64) (int, error) {
	if q == nil || q.cache == nil {
		return 0, errors.New("email queue: cache is nil")
	}

	maxScore := fmt.Sprintf("%d", time.Now().Unix())
	items, err := q.cache.ZRangeByScore(q.retryKey, "-inf", maxScore, limit)
	if err != nil {
		return 0, err
	}

	if len(items) == 0 {
		return 0, nil
	}

	var moved int
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, err := q.cache.ZRem(q.retryKey, item); err != nil {
			return moved, err
		}
		if _, err := q.cache.LPush(q.pendingKey, item); err != nil {
			return moved, err
		}
		moved++
	}

	return moved, nil
}

func (q *RedisQueue) RecoverInFlight(ctx context.Context) (int, error) {
	if q == nil || q.cache == nil {
		return 0, errors.New("email queue: cache is nil")
	}

	items, err := q.cache.LRange(q.processingKey, 0, -1)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, nil
	}

	for _, item := range items {
		if item == "" {
			continue
		}
		if _, err := q.cache.LPush(q.pendingKey, item); err != nil {
			return 0, err
		}
	}
	if err := q.cache.Delete(q.processingKey); err != nil {
		return 0, err
	}

	return len(items), nil
}
