package rbac

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	rbacdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/rbac"
	rbacusecase "github.com/mzulfanw/boilerplate-go-fiber/internal/service/rbac"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/validation"
)

type Handler struct {
	service *rbacusecase.Service
}

func NewHandler(service *rbacusecase.Service) *Handler {
	return &Handler{service: service}
}

// ListRoles godoc
// @Summary List roles
// @Tags RBAC
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]RoleResponse}
// @Failure 401 {object} response.Response
// @Router /rbac/roles [get]
func (h *Handler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.service.ListRoles(c.UserContext())
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapRoles(roles),
	}
	return c.Status(resp.Code).JSON(resp)
}

// GetRole godoc
// @Summary Get role
// @Tags RBAC
// @Security BearerAuth
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} response.Response{data=RoleResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/roles/{id} [get]
func (h *Handler) GetRole(c *fiber.Ctx) error {
	roleID, err := validation.RequireParam(c.Params("id"), "role id")
	if err != nil {
		return err
	}

	role, err := h.service.GetRole(c.UserContext(), roleID)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapRole(role),
	}
	return c.Status(resp.Code).JSON(resp)
}

// CreateRole godoc
// @Summary Create role
// @Tags RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body RoleRequest true "Create role payload"
// @Success 201 {object} response.Response{data=RoleResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /rbac/roles [post]
func (h *Handler) CreateRole(c *fiber.Ctx) error {
	var req RoleRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)

	role, err := h.service.CreateRole(c.UserContext(), req.Name, req.Description)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusCreated,
		Message: "created",
		Data:    mapRole(role),
	}
	return c.Status(resp.Code).JSON(resp)
}

// UpdateRole godoc
// @Summary Update role
// @Tags RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param payload body RoleRequest true "Update role payload"
// @Success 200 {object} response.Response{data=RoleResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/roles/{id} [put]
func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	roleID, err := validation.RequireParam(c.Params("id"), "role id")
	if err != nil {
		return err
	}

	var req RoleRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)

	role, err := h.service.UpdateRole(c.UserContext(), roleID, req.Name, req.Description)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapRole(role),
	}
	return c.Status(resp.Code).JSON(resp)
}

// DeleteRole godoc
// @Summary Delete role
// @Tags RBAC
// @Security BearerAuth
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/roles/{id} [delete]
func (h *Handler) DeleteRole(c *fiber.Ctx) error {
	roleID, err := validation.RequireParam(c.Params("id"), "role id")
	if err != nil {
		return err
	}

	if err := h.service.DeleteRole(c.UserContext(), roleID); err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
	}
	return c.Status(resp.Code).JSON(resp)
}

// ListPermissions godoc
// @Summary List permissions
// @Tags RBAC
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]PermissionResponse}
// @Failure 401 {object} response.Response
// @Router /rbac/permissions [get]
func (h *Handler) ListPermissions(c *fiber.Ctx) error {
	permissions, err := h.service.ListPermissions(c.UserContext())
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapPermissions(permissions),
	}
	return c.Status(resp.Code).JSON(resp)
}

// GetPermission godoc
// @Summary Get permission
// @Tags RBAC
// @Security BearerAuth
// @Produce json
// @Param id path string true "Permission ID"
// @Success 200 {object} response.Response{data=PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/permissions/{id} [get]
func (h *Handler) GetPermission(c *fiber.Ctx) error {
	permissionID, err := validation.RequireParam(c.Params("id"), "permission id")
	if err != nil {
		return err
	}

	permission, err := h.service.GetPermission(c.UserContext(), permissionID)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapPermission(permission),
	}
	return c.Status(resp.Code).JSON(resp)
}

// CreatePermission godoc
// @Summary Create permission
// @Tags RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body PermissionRequest true "Create permission payload"
// @Success 201 {object} response.Response{data=PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /rbac/permissions [post]
func (h *Handler) CreatePermission(c *fiber.Ctx) error {
	var req PermissionRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)

	permission, err := h.service.CreatePermission(c.UserContext(), req.Name, req.Description)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusCreated,
		Message: "created",
		Data:    mapPermission(permission),
	}
	return c.Status(resp.Code).JSON(resp)
}

// UpdatePermission godoc
// @Summary Update permission
// @Tags RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Param payload body PermissionRequest true "Update permission payload"
// @Success 200 {object} response.Response{data=PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/permissions/{id} [put]
func (h *Handler) UpdatePermission(c *fiber.Ctx) error {
	permissionID, err := validation.RequireParam(c.Params("id"), "permission id")
	if err != nil {
		return err
	}

	var req PermissionRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)

	permission, err := h.service.UpdatePermission(c.UserContext(), permissionID, req.Name, req.Description)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapPermission(permission),
	}
	return c.Status(resp.Code).JSON(resp)
}

// DeletePermission godoc
// @Summary Delete permission
// @Tags RBAC
// @Security BearerAuth
// @Produce json
// @Param id path string true "Permission ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/permissions/{id} [delete]
func (h *Handler) DeletePermission(c *fiber.Ctx) error {
	permissionID, err := validation.RequireParam(c.Params("id"), "permission id")
	if err != nil {
		return err
	}

	if err := h.service.DeletePermission(c.UserContext(), permissionID); err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
	}
	return c.Status(resp.Code).JSON(resp)
}

// ListRolePermissions godoc
// @Summary List role permissions
// @Tags RBAC
// @Security BearerAuth
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} response.Response{data=RolePermissionsResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/roles/{id}/permissions [get]
func (h *Handler) ListRolePermissions(c *fiber.Ctx) error {
	roleID, err := validation.RequireParam(c.Params("id"), "role id")
	if err != nil {
		return err
	}

	permissions, err := h.service.ListRolePermissions(c.UserContext(), roleID)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data: RolePermissionsResponse{
			RoleID:      roleID,
			Permissions: mapPermissions(permissions),
		},
	}
	return c.Status(resp.Code).JSON(resp)
}

// UpdateRolePermissions godoc
// @Summary Replace role permissions
// @Tags RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param payload body RolePermissionsRequest true "Role permissions payload"
// @Success 200 {object} response.Response{data=RolePermissionsResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /rbac/roles/{id}/permissions [put]
func (h *Handler) UpdateRolePermissions(c *fiber.Ctx) error {
	roleID, err := validation.RequireParam(c.Params("id"), "role id")
	if err != nil {
		return err
	}

	var req RolePermissionsRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.service.ReplaceRolePermissions(c.UserContext(), roleID, req.PermissionIDs); err != nil {
		return mapRBACError(err)
	}

	permissions, err := h.service.ListRolePermissions(c.UserContext(), roleID)
	if err != nil {
		return mapRBACError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data: RolePermissionsResponse{
			RoleID:      roleID,
			Permissions: mapPermissions(permissions),
		},
	}
	return c.Status(resp.Code).JSON(resp)
}

func mapRBACError(err error) error {
	switch {
	case errors.Is(err, rbacdomain.ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	case errors.Is(err, rbacdomain.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "resource not found")
	case errors.Is(err, rbacdomain.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, "resource already exists")
	default:
		return err
	}
}

func mapRoles(roles []rbacdomain.Role) []RoleResponse {
	result := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, mapRole(role))
	}
	return result
}

func mapRole(role rbacdomain.Role) RoleResponse {
	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func mapPermissions(permissions []rbacdomain.Permission) []PermissionResponse {
	result := make([]PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, mapPermission(permission))
	}
	return result
}

func mapPermission(permission rbacdomain.Permission) PermissionResponse {
	return PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt.UTC().Format(time.RFC3339),
	}
}
