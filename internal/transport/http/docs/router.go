package apidocs

import (
	_ "embed"
	"strings"

	"github.com/gofiber/fiber/v2"
	swagdocs "github.com/mzulfanw/boilerplate-go-fiber/docs"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
)

const (
	defaultSwaggerPath = "/docs"
	specSuffix         = "/doc.json"
)

var (
	//go:embed swagger.html
	swaggerHTML string
)

type Router struct {
	cfg config.Config
}

func NewRouter(cfg config.Config) *Router {
	return &Router{cfg: cfg}
}

func (r *Router) Register(app *fiber.App) {
	if !r.cfg.SwaggerEnabled {
		return
	}

	base := strings.TrimSpace(r.cfg.SwaggerPath)
	if base == "" {
		base = defaultSwaggerPath
	}
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	base = strings.TrimRight(base, "/")
	if base == "" {
		base = "/"
	}
	specPath := base + specSuffix

	serveUI := func(c *fiber.Ctx) error {
		html := strings.ReplaceAll(swaggerHTML, "{{SPEC_URL}}", specPath)
		return c.Type("html").SendString(html)
	}

	app.Get(base, serveUI)
	if base != "/" {
		app.Get(base+"/", serveUI)
	}
	app.Get(specPath, func(c *fiber.Ctx) error {
		return c.Type("json").SendString(swagdocs.SwaggerInfo.ReadDoc())
	})
}
