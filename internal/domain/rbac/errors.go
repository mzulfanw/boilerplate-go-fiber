package rbac

import "errors"

var (
	ErrNotFound     = errors.New("rbac: not found")
	ErrConflict     = errors.New("rbac: conflict")
	ErrInvalidInput = errors.New("rbac: invalid input")
)
