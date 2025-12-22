package rbac

import (
	"time"

	"github.com/mzulfanw/boilerplate-go-fiber/internal/domain/query"
)

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

type ListFilterRole struct {
	Search      string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Pagination  query.Pagination
}

type ListRole struct {
	Role  []Role
	Total int
}

type ListPermission struct {
	Permission []Permission
	Total      int
}

type ListFilterPermission struct {
	Search      string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Pagination  query.Pagination
}
