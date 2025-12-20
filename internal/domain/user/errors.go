package user

import "errors"

var (
	ErrNotFound     = errors.New("user: not found")
	ErrConflict     = errors.New("user: conflict")
	ErrInvalidInput = errors.New("user: invalid input")
)
