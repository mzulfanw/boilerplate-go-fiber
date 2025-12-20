package user

import "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/query"

type ListFilter struct {
	Search     string
	IsActive   *bool
	Pagination query.Pagination
}

type ListResult struct {
	Users []User
	Total int
}
