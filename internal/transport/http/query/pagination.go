package query

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	domainquery "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/query"
)

const (
	defaultPage    = 1
	defaultPerPage = 10
	maxPerPage     = 100
)

func ParsePagination(c *fiber.Ctx) (domainquery.Pagination, error) {
	page, err := parsePositiveInt(c.Query("page"), "page", defaultPage)
	if err != nil {
		return domainquery.Pagination{}, err
	}

	perPage, err := parsePositiveInt(c.Query("per_page"), "per_page", defaultPerPage)
	if err != nil {
		return domainquery.Pagination{}, err
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return domainquery.Pagination{
		Page:    page,
		PerPage: perPage,
	}, nil
}

func ParseSearch(c *fiber.Ctx, key string) string {
	return strings.TrimSpace(c.Query(key))
}

func ParseOptionalBool(c *fiber.Ctx, key string) (*bool, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, key+" must be a boolean")
	}

	return &parsed, nil
}

func parsePositiveInt(raw, label string, fallback int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, label+" must be a positive integer")
	}

	return parsed, nil
}
