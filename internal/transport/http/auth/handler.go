package auth

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	authusecase "github.com/mzulfanw/boilerplate-go-fiber/internal/service/auth"
	emailservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/email"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/validation"
)

type Handler struct {
	service  *authusecase.Service
	email    *emailservice.Service
	renderer *emailservice.Renderer
	resetURL string
	appName  string
}

func NewHandler(service *authusecase.Service, emailService *emailservice.Service, renderer *emailservice.Renderer, resetURL, appName string) *Handler {
	return &Handler{
		service:  service,
		email:    emailService,
		renderer: renderer,
		resetURL: strings.TrimSpace(resetURL),
		appName:  strings.TrimSpace(appName),
	}
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
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}
	req.Email = strings.TrimSpace(req.Email)

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
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)

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
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)

	if err := h.service.Logout(c.UserContext(), req.RefreshToken); err != nil {
		return err
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
	}
	return c.Status(resp.Code).JSON(resp)
}

// ForgotPassword godoc
// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body ForgotPasswordRequest true "Forgot password payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.RequestPasswordReset(c.UserContext(), req.Email)
	if err != nil {
		return mapAuthError(err)
	}

	if result.ShouldSend {
		if h.email == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "email service is not configured")
		}
		resetLink := buildResetLink(h.resetURL, result.Token)
		subject := "Reset your password"
		body := fmt.Sprintf("We received a request to reset your password.\n\nReset link: %s\n\nThis link expires at %s.\nIf you did not request a reset, you can ignore this email.",
			resetLink,
			result.ExpiresAt.Format(time.RFC1123),
		)
		contentType := "text/plain; charset=utf-8"

		if h.renderer != nil {
			rendered, err := h.renderer.Render("reset_password", emailservice.TemplateData{
				AppName:        h.appName,
				RecipientEmail: result.Email,
				ActionURL:      resetLink,
				ActionLabel:    "Reset Password",
				ExpiresAt:      result.ExpiresAt.Format(time.RFC1123),
			})
			if err != nil {
				return err
			}
			if strings.TrimSpace(rendered.Subject) != "" {
				subject = rendered.Subject
			}
			if strings.TrimSpace(rendered.HTML) != "" {
				body = rendered.HTML
				contentType = "text/html; charset=utf-8"
			} else if strings.TrimSpace(rendered.Text) != "" {
				body = rendered.Text
			}
		}
		if err := h.email.Enqueue(c.UserContext(), emailservice.Message{
			To:          []string{result.Email},
			Subject:     subject,
			Body:        body,
			ContentType: contentType,
		}); err != nil {
			return err
		}
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
	}
	return c.Status(resp.Code).JSON(resp)
}

// ResetPassword godoc
// @Summary Reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body ResetPasswordRequest true "Reset password payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.service.ResetPassword(c.UserContext(), req.Token, req.Password); err != nil {
		return mapAuthError(err)
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
	case errors.Is(err, authusecase.ErrInvalidResetToken):
		return fiber.NewError(fiber.StatusBadRequest, "invalid reset token")
	case errors.Is(err, authusecase.ErrInvalidPassword):
		return fiber.NewError(fiber.StatusBadRequest, "invalid password")
	default:
		return err
	}
}

func buildResetLink(baseURL, token string) string {
	escapedToken := url.QueryEscape(token)
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return escapedToken
	}
	if strings.Contains(trimmed, "%s") {
		return fmt.Sprintf(trimmed, escapedToken)
	}
	if strings.Contains(trimmed, "?") {
		return trimmed + "&token=" + escapedToken
	}
	return trimmed + "?token=" + escapedToken
}
