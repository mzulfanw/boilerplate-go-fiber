package user

import (
	"github.com/gofiber/fiber/v2"
	httptransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http"
)

const (
	permUserRead       = "user.read"
	permUserCreate     = "user.create"
	permUserUpdate     = "user.update"
	permUserDelete     = "user.delete"
	permUserRoleRead   = "user.role.read"
	permUserRoleUpdate = "user.role.update"
)

type Router struct {
	handler *Handler
	auth    *httptransport.AuthMiddleware
}

func NewRouter(handler *Handler, auth *httptransport.AuthMiddleware) *Router {
	return &Router{handler: handler, auth: auth}
}

func (r *Router) Register(app *fiber.App) {
	if r == nil || r.handler == nil || r.auth == nil || app == nil {
		return
	}

	group := app.Group("/users", r.auth.RequireAuth())
	group.Get("/", r.auth.RequirePermissions(permUserRead), r.handler.ListUsers)
	group.Get("/:id", r.auth.RequirePermissions(permUserRead), r.handler.GetUser)
	group.Post("/", r.auth.RequirePermissions(permUserCreate), r.handler.CreateUser)
	group.Put("/:id", r.auth.RequirePermissions(permUserUpdate), r.handler.UpdateUser)
	group.Delete("/:id", r.auth.RequirePermissions(permUserDelete), r.handler.DeleteUser)
	group.Get("/:id/roles", r.auth.RequirePermissions(permUserRoleRead), r.handler.ListUserRoles)
	group.Put("/:id/roles", r.auth.RequirePermissions(permUserRoleUpdate), r.handler.UpdateUserRoles)
}
