package rbac

import (
	"github.com/gofiber/fiber/v2"
	httptransport "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http"
)

const (
	permRoleRead             = "role.read"
	permRoleCreate           = "role.create"
	permRoleUpdate           = "role.update"
	permRoleDelete           = "role.delete"
	permPermissionRead       = "permission.read"
	permPermissionCreate     = "permission.create"
	permPermissionUpdate     = "permission.update"
	permPermissionDelete     = "permission.delete"
	permRolePermissionRead   = "role.permission.read"
	permRolePermissionUpdate = "role.permission.update"
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

	group := app.Group("/rbac", r.auth.RequireAuth())

	group.Get("/roles", r.auth.RequirePermissions(permRoleRead), r.handler.ListRoles)
	group.Get("/roles/:id", r.auth.RequirePermissions(permRoleRead), r.handler.GetRole)
	group.Post("/roles", r.auth.RequirePermissions(permRoleCreate), r.handler.CreateRole)
	group.Put("/roles/:id", r.auth.RequirePermissions(permRoleUpdate), r.handler.UpdateRole)
	group.Delete("/roles/:id", r.auth.RequirePermissions(permRoleDelete), r.handler.DeleteRole)

	group.Get("/roles/:id/permissions", r.auth.RequirePermissions(permRolePermissionRead), r.handler.ListRolePermissions)
	group.Put("/roles/:id/permissions", r.auth.RequirePermissions(permRolePermissionUpdate), r.handler.UpdateRolePermissions)

	group.Get("/permissions", r.auth.RequirePermissions(permPermissionRead), r.handler.ListPermissions)
	group.Get("/permissions/:id", r.auth.RequirePermissions(permPermissionRead), r.handler.GetPermission)
	group.Post("/permissions", r.auth.RequirePermissions(permPermissionCreate), r.handler.CreatePermission)
	group.Put("/permissions/:id", r.auth.RequirePermissions(permPermissionUpdate), r.handler.UpdatePermission)
	group.Delete("/permissions/:id", r.auth.RequirePermissions(permPermissionDelete), r.handler.DeletePermission)
}
