package query

type Pagination struct {
	Page    int
	PerPage int
}

func (p Pagination) Limit() int {
	if p.PerPage <= 0 {
		return 0
	}
	return p.PerPage
}

func (p Pagination) Offset() int {
	if p.Page <= 1 || p.PerPage <= 0 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}
