package email

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type Service struct {
	queue Queue
}

func NewService(queue Queue) *Service {
	return &Service{queue: queue}
}

func (s *Service) Enqueue(ctx context.Context, msg Message) error {
	if s == nil || s.queue == nil {
		return errors.New("email service: queue is nil")
	}
	if len(msg.To) == 0 {
		return errors.New("email service: recipient is empty")
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return errors.New("email service: subject is empty")
	}
	if strings.TrimSpace(msg.Body) == "" {
		return errors.New("email service: body is empty")
	}

	if strings.TrimSpace(msg.ContentType) == "" {
		msg.ContentType = "text/plain; charset=utf-8"
	}

	job := Job{
		ID:        newJobID(),
		Message:   msg,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return s.queue.Enqueue(ctx, job)
}

func newJobID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(buf)
}
