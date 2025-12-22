package payment

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/xendit/xendit-go/v7/invoice"

	paymentdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/payment"
	paymentusecase "github.com/mzulfanw/boilerplate-go-fiber/internal/service/payment"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/validation"
)

type Handler struct {
	service      *paymentusecase.Service
	webhookToken string
}

func NewHandler(service *paymentusecase.Service, webhookToken string) *Handler {
	return &Handler{
		service:      service,
		webhookToken: strings.TrimSpace(webhookToken),
	}
}

// Check Balance godoc
// @Summary Check Balance
// @Tags Payment
// @Produce json
// @Success 200 {object} response.Response{data=BalanceResponse}
// @Failure 500 {object} response.Response
// @Router /payment/balance [get]
func (h *Handler) CheckBalance(c *fiber.Ctx) error {
	balance, err := h.service.CheckBalance(c.UserContext())
	if err != nil {
		return mapPaymentError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data: BalanceResponse{
			Balance: balance,
		},
	}
	return c.Status(resp.Code).JSON(resp)
}

// Get Invoice By Id godoc
// @Summary Get Invoice By Id
// @Tags Payment
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.Response{data=invoice.Invoice}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /payment/invoice/{id} [get]
func (h *Handler) GetInvoiceById(c *fiber.Ctx) error {
	invoiceID, err := validation.RequireParam(c.Params("id"), "invoice id")
	if err != nil {
		return err
	}

	invoice, err := h.service.GetInvoiceById(c.UserContext(), invoiceID)
	if err != nil {
		return mapPaymentError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    invoice,
	}
	return c.Status(resp.Code).JSON(resp)
}

// Create Invoice godoc
// @Summary Create Invoice
// @Tags Payment
// @Accept json
// @Produce json
// @Param Idempotency-Key header string false "Idempotency key"
// @Param payload body CreateInvoiceRequest true "Create invoice payload"
// @Success 201 {object} response.Response{data=invoice.Invoice}
// @Success 200 {object} response.Response{data=invoice.Invoice}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /payment/invoice [post]
func (h *Handler) CreateInvoice(c *fiber.Ctx) error {
	var req CreateInvoiceRequest
	if err := validation.ParseAndValidate(c, &req); err != nil {
		return err
	}

	input := paymentdomain.CreateInvoiceInput{
		IdempotencyKey:               getIdempotencyKey(c),
		ExternalID:                   req.ExternalID,
		Amount:                       req.Amount,
		PayerEmail:                   req.PayerEmail,
		Description:                  req.Description,
		InvoiceDurationSeconds:       req.InvoiceDurationSeconds,
		CallbackVirtualAccountID:     req.CallbackVirtualAccountID,
		ShouldSendEmail:              req.ShouldSendEmail,
		SuccessRedirectURL:           req.SuccessRedirectURL,
		FailureRedirectURL:           req.FailureRedirectURL,
		PaymentMethods:               req.PaymentMethods,
		MidLabel:                     req.MidLabel,
		ShouldAuthenticateCreditCard: req.ShouldAuthenticateCreditCard,
		Currency:                     req.Currency,
		ReminderTimeSeconds:          req.ReminderTimeSeconds,
		Locale:                       req.Locale,
		ReminderTimeUnit:             req.ReminderTimeUnit,
		Items:                        mapInvoiceItems(req.Items),
		Fees:                         mapInvoiceFees(req.Fees),
		Metadata:                     req.Metadata,
	}

	invoice, reused, err := h.service.CreateInvoice(c.UserContext(), input)
	if err != nil {
		return mapPaymentError(err)
	}

	statusCode := fiber.StatusCreated
	message := "created"
	if reused {
		statusCode = fiber.StatusOK
		message = "ok"
	}

	resp := response.Response{
		Code:    statusCode,
		Message: message,
		Data:    invoice,
	}
	return c.Status(resp.Code).JSON(resp)
}

// Expire Invoice godoc
// @Summary Expire Invoice
// @Tags Payment
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.Response{data=invoice.Invoice}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /payment/invoice/{id}/expire [post]
func (h *Handler) ExpireInvoice(c *fiber.Ctx) error {
	invoiceID, err := validation.RequireParam(c.Params("id"), "invoice id")
	if err != nil {
		return err
	}

	invoice, err := h.service.ExpireInvoice(c.UserContext(), invoiceID)
	if err != nil {
		return mapPaymentError(err)
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
		Data:    invoice,
	}
	return c.Status(resp.Code).JSON(resp)
}

// Invoice Webhook godoc
// @Summary Xendit invoice webhook
// @Tags Payment
// @Accept json
// @Produce json
// @Param payload body invoice.InvoiceCallback true "Invoice callback payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /payment/webhook/xendit/invoice [post]
func (h *Handler) InvoiceWebhook(c *fiber.Ctx) error {
	if err := h.verifyWebhookToken(c); err != nil {
		return err
	}

	payloadBytes := c.Body()
	if len(payloadBytes) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "empty request body")
	}

	var payload invoice.InvoiceCallback
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	payloadHash := hashPayload(payloadBytes)
	fields := logrus.Fields{
		"invoice_id":   payload.Id,
		"external_id":  payload.ExternalId,
		"status":       payload.Status,
		"amount":       payload.Amount,
		"payload":      string(payloadBytes),
		"payload_hash": payloadHash,
		"request_id":   c.Get(fiber.HeaderXRequestID),
	}
	logrus.WithFields(fields).Info("xendit invoice webhook received")

	if err := h.handleInvoiceWebhookWithRetry(c.UserContext(), payload, payloadHash); err != nil {
		if errors.Is(err, paymentdomain.ErrInvalidInput) {
			logrus.WithFields(fields).WithError(err).Warn("xendit invoice webhook invalid payload")
			return mapPaymentError(err)
		}
		logrus.WithFields(fields).WithError(err).Error("xendit invoice webhook processing failed")
		return fiber.NewError(fiber.StatusInternalServerError, "webhook processing failed")
	}

	resp := response.Response{
		Code:    fiber.StatusOK,
		Message: "ok",
	}
	return c.Status(resp.Code).JSON(resp)
}

func mapInvoiceItems(items []InvoiceItemRequest) []paymentdomain.InvoiceItem {
	if len(items) == 0 {
		return nil
	}
	result := make([]paymentdomain.InvoiceItem, 0, len(items))
	for _, item := range items {
		result = append(result, paymentdomain.InvoiceItem{
			Name:        item.Name,
			Price:       item.Price,
			Quantity:    item.Quantity,
			ReferenceID: item.ReferenceID,
			URL:         item.URL,
			Category:    item.Category,
		})
	}
	return result
}

func mapInvoiceFees(fees []InvoiceFeeRequest) []paymentdomain.InvoiceFee {
	if len(fees) == 0 {
		return nil
	}
	result := make([]paymentdomain.InvoiceFee, 0, len(fees))
	for _, fee := range fees {
		result = append(result, paymentdomain.InvoiceFee{
			Type:  fee.Type,
			Value: fee.Value,
		})
	}
	return result
}

func mapPaymentError(err error) error {
	if errors.Is(err, paymentdomain.ErrInvalidInput) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}
	if errors.Is(err, paymentdomain.ErrIdempotencyInProgress) {
		return fiber.NewError(fiber.StatusConflict, "request is still being processed")
	}
	return err
}

func (h *Handler) verifyWebhookToken(c *fiber.Ctx) error {
	if h == nil || h.webhookToken == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "webhook token not configured")
	}

	token := strings.TrimSpace(c.Get("x-callback-token"))
	if token == "" {
		token = strings.TrimSpace(c.Get("xendit-callback-token"))
	}
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing callback token")
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(h.webhookToken)) != 1 {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid callback token")
	}

	return nil
}

func getIdempotencyKey(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}
	key := strings.TrimSpace(c.Get("Idempotency-Key"))
	if key == "" {
		key = strings.TrimSpace(c.Get("X-Idempotency-Key"))
	}
	return key
}

func (h *Handler) handleInvoiceWebhookWithRetry(ctx context.Context, payload invoice.InvoiceCallback, payloadHash string) error {
	const maxAttempts = 3
	const baseDelay = 200 * time.Millisecond

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := h.service.HandleInvoiceWebhook(ctx, payload, payloadHash); err != nil {
			if errors.Is(err, paymentdomain.ErrInvalidInput) {
				return err
			}
			lastErr = err
			if attempt < maxAttempts {
				time.Sleep(baseDelay * time.Duration(attempt))
				continue
			}
			return err
		}
		return nil
	}

	return lastErr
}

func hashPayload(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
