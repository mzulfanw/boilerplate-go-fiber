package email

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type WorkerOptions struct {
	ReserveTimeout   time.Duration
	RequeueInterval  time.Duration
	SendTimeout      time.Duration
	MaxAttempts      int
	RetryDelays      []time.Duration
	RequeueBatchSize int64
	RecoverInFlight  bool
}

type Worker struct {
	queue  Queue
	sender Sender
	opts   WorkerOptions
}

func NewWorker(queue Queue, sender Sender, opts WorkerOptions) *Worker {
	if opts.ReserveTimeout <= 0 {
		opts.ReserveTimeout = 5 * time.Second
	}
	if opts.RequeueInterval <= 0 {
		opts.RequeueInterval = 5 * time.Second
	}
	if opts.SendTimeout <= 0 {
		opts.SendTimeout = 15 * time.Second
	}
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 5
	}
	if len(opts.RetryDelays) == 0 {
		opts.RetryDelays = []time.Duration{
			10 * time.Second,
			30 * time.Second,
			1 * time.Minute,
			5 * time.Minute,
		}
	}
	if opts.RequeueBatchSize <= 0 {
		opts.RequeueBatchSize = 50
	}

	return &Worker{
		queue:  queue,
		sender: sender,
		opts:   opts,
	}
}

func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.queue == nil || w.sender == nil {
		return
	}

	if w.opts.RecoverInFlight {
		if moved, err := w.queue.RecoverInFlight(ctx); err != nil {
			logrus.WithError(err).Warn("email worker recover in-flight failed")
		} else if moved > 0 {
			logrus.WithField("count", moved).Info("email worker requeued in-flight jobs")
		}
	}

	go w.requeueLoop(ctx)

	for {
		if ctx.Err() != nil {
			return
		}

		job, err := w.queue.Reserve(ctx, w.opts.ReserveTimeout)
		if err != nil {
			if errors.Is(err, ErrQueueEmpty) {
				continue
			}
			if errors.Is(err, ErrInvalidPayload) {
				if dlqErr := w.queue.DeadLetter(ctx, job, "invalid payload"); dlqErr != nil {
					logrus.WithError(dlqErr).Warn("email worker dead-letter failed")
				}
				continue
			}
			logrus.WithError(err).Warn("email worker reserve failed")
			continue
		}

		w.handleJob(ctx, job)
	}
}

func (w *Worker) handleJob(ctx context.Context, job Job) {
	sendCtx, cancel := context.WithTimeout(ctx, w.opts.SendTimeout)
	err := w.sender.Send(sendCtx, job.Message)
	cancel()

	if err == nil {
		if ackErr := w.queue.Ack(ctx, job); ackErr != nil {
			logrus.WithError(ackErr).Warn("email worker ack failed")
		}
		return
	}

	job.Attempts++
	job.LastError = err.Error()
	job.UpdatedAt = time.Now().UTC()

	maxAttempts := job.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = w.opts.MaxAttempts
	}
	if job.Attempts >= maxAttempts {
		if dlqErr := w.queue.DeadLetter(ctx, job, err.Error()); dlqErr != nil {
			logrus.WithError(dlqErr).Warn("email worker dead-letter failed")
		}
		logrus.WithFields(logrus.Fields{
			"job_id":   job.ID,
			"attempts": job.Attempts,
			"error":    err.Error(),
		}).Warn("email worker exceeded retry attempts")
		return
	}

	delay := w.retryDelay(job.Attempts)
	if retryErr := w.queue.Retry(ctx, job, delay); retryErr != nil {
		logrus.WithError(retryErr).Warn("email worker retry scheduling failed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"attempts": job.Attempts,
		"delay":    delay.String(),
		"error":    err.Error(),
	}).Warn("email worker retry scheduled")
}

func (w *Worker) retryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return w.opts.RetryDelays[0]
	}
	if attempt <= len(w.opts.RetryDelays) {
		return w.opts.RetryDelays[attempt-1]
	}

	last := w.opts.RetryDelays[len(w.opts.RetryDelays)-1]
	return last + time.Duration(attempt-len(w.opts.RetryDelays))*time.Minute
}

func (w *Worker) requeueLoop(ctx context.Context) {
	ticker := time.NewTicker(w.opts.RequeueInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
			if _, err := w.queue.RequeueDue(ctx, w.opts.RequeueBatchSize); err != nil {
				logrus.WithError(err).Warn("email worker requeue failed")
			}
		}
	}
}
