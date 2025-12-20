package rbac

import "time"

type Role struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

type Permission struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}
