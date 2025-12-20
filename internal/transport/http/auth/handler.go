package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	authusecase "github.com/mzulfanw/boilerplate-go-fiber/internal/service/auth"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
)

type Handler struct {
	service *authusecase.Service
}

func NewHandler(service *authusecase.Service) *Handler {
	return &Handler{service: service}
}

// Login godoc
// @Summary Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body LoginRequest true "Login payload"
// @Success 200 {object} response.Response{data=TokenResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 423 {object} response.Response
// @Router /auth/login [post]
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	ip := c.IP()
	ua := c.Get(fiber.HeaderUserAgent)
	result, err := h.service.Login(c.UserContext(), req.Email, req.Password, ip, ua)
	if err != nil {
		return mapAuthError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data: TokenResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			TokenType:    result.TokenType,
			ExpiresIn:    result.ExpiresIn,
		},
	}
	return c.Status(resp.Code).JSON(resp)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body RefreshRequest true "Refresh token payload"
// @Success 200 {object} response.Response{data=TokenResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "refresh_token is required")
	}

	ip := c.IP()
	ua := c.Get(fiber.HeaderUserAgent)
	result, err := h.service.Refresh(c.UserContext(), req.RefreshToken, ip, ua)
	if err != nil {
		return mapAuthError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data: TokenResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			TokenType:    result.TokenType,
			ExpiresIn:    result.ExpiresIn,
		},
	}
	return c.Status(resp.Code).JSON(resp)
}

// Logout godoc
// @Summary Logout
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body LogoutRequest true "Logout payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /auth/logout [post]
func (h *Handler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "refresh_token is required")
	}

	if err := h.service.Logout(c.UserContext(), req.RefreshToken); err != nil {
		return err
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
	}
	return c.Status(resp.Code).JSON(resp)
}

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, authusecase.ErrInvalidCredentials):
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, authusecase.ErrInvalidToken):
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	case errors.Is(err, authusecase.ErrUserDisabled):
		return fiber.NewError(fiber.StatusForbidden, "user is disabled")
	case errors.Is(err, authusecase.ErrUserLocked):
		return fiber.NewError(fiber.StatusLocked, "user is locked")
	default:
		return err
	}
}
