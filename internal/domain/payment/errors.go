package payment

import "errors"

var (
	ErrInvalidInput          = errors.New("payment: invalid input")
	ErrIdempotencyInProgress = errors.New("payment: idempotency in progress")
)
