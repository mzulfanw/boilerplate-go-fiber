package user

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	rbacdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/rbac"
	userdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/user"
	userusecase "github.com/mzulfanw/boilerplate-go-fiber/internal/service/user"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
)

type Handler struct {
	service *userusecase.Service
}

func NewHandler(service *userusecase.Service) *Handler {
	return &Handler{service: service}
}

// ListUsers godoc
// @Summary List users
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]UserResponse}
// @Failure 401 {object} response.Response
// @Router /users [get]
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.service.ListUsers(c.UserContext())
	if err != nil {
		return mapUserError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapUsers(users),
	}
	return c.Status(resp.Code).JSON(resp)
}

// GetUser godoc
// @Summary Get user by id
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [get]
func (h *Handler) GetUser(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("id"))
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user id is required")
	}

	user, err := h.service.GetUser(c.UserContext(), userID)
	if err != nil {
		return mapUserError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapUser(user),
	}
	return c.Status(resp.Code).JSON(resp)
}

// CreateUser godoc
// @Summary Create user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateUserRequest true "Create user payload"
// @Success 201 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /users [post]
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || strings.TrimSpace(req.Password) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	user, err := h.service.CreateUser(c.UserContext(), req.Email, req.Password, req.IsActive, req.RoleIDs)
	if err != nil {
		return mapUserError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusCreated,
		Message: "created",
		Data:    mapUser(user),
	}
	return c.Status(resp.Code).JSON(resp)
}

// UpdateUser godoc
// @Summary Update user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body UpdateUserRequest true "Update user payload"
// @Success 200 {object} response.Response{data=UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [put]
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("id"))
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user id is required")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.Email == nil && req.Password == nil && req.IsActive == nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fields to update")
	}

	if req.Email != nil {
		value := strings.TrimSpace(*req.Email)
		if value == "" {
			return fiber.NewError(fiber.StatusBadRequest, "email cannot be empty")
		}
		req.Email = &value
	}

	if req.Password != nil && strings.TrimSpace(*req.Password) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "password cannot be empty")
	}

	user, err := h.service.UpdateUser(c.UserContext(), userID, req.Email, req.Password, req.IsActive)
	if err != nil {
		return mapUserError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    mapUser(user),
	}
	return c.Status(resp.Code).JSON(resp)
}

// DeleteUser godoc
// @Summary Delete user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [delete]
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("id"))
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user id is required")
	}

	if err := h.service.DeleteUser(c.UserContext(), userID); err != nil {
		return mapUserError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
	}
	return c.Status(resp.Code).JSON(resp)
}

// ListUserRoles godoc
// @Summary List user roles
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=UserRolesResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id}/roles [get]
func (h *Handler) ListUserRoles(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("id"))
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user id is required")
	}

	roles, err := h.service.ListUserRoles(c.UserContext(), userID)
	if err != nil {
		return mapUserError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data: UserRolesResponse{
			UserID: userID,
			Roles:  mapRoles(roles),
		},
	}
	return c.Status(resp.Code).JSON(resp)
}

// UpdateUserRoles godoc
// @Summary Replace user roles
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body UserRolesRequest true "Role ids"
// @Success 200 {object} response.Response{data=UserRolesResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id}/roles [put]
func (h *Handler) UpdateUserRoles(c *fiber.Ctx) error {
	userID := strings.TrimSpace(c.Params("id"))
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user id is required")
	}

	var req UserRolesRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.service.ReplaceUserRoles(c.UserContext(), userID, req.RoleIDs); err != nil {
		return mapUserError(err)
	}

	roles, err := h.service.ListUserRoles(c.UserContext(), userID)
	if err != nil {
		return mapUserError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data: UserRolesResponse{
			UserID: userID,
			Roles:  mapRoles(roles),
		},
	}
	return c.Status(resp.Code).JSON(resp)
}

func mapUserError(err error) error {
	switch {
	case errors.Is(err, userdomain.ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	case errors.Is(err, userdomain.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	case errors.Is(err, userdomain.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, "user already exists")
	default:
		return err
	}
}

func mapUsers(users []userdomain.User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, mapUser(user))
	}
	return result
}

func mapUser(user userdomain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.UTC().Format(time.RFC3339),
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
