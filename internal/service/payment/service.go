package payment

import (
	"context"
	"errors"
	"strings"
	"time"

	redisinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/redis"
	xenditinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/xendit"
	paymentdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/payment"
	"github.com/sirupsen/logrus"
	"github.com/xendit/xendit-go/v7/invoice"
)

const (
	idempotencyPendingValue = "__pending__"
	idempotencyTTL          = 24 * time.Hour
	idempotencyPendingTTL   = 2 * time.Minute
	webhookDedupTTL         = 7 * 24 * time.Hour
)

type Service struct {
	xenditClient xenditinfra.XenditClient
	cache        redisinfra.Cache
}

func NewService(xenditClient xenditinfra.XenditClient, cache redisinfra.Cache) (*Service, error) {
	if xenditClient == nil {
		return nil, errors.New("payment: xendit client is nil")
	}
	return &Service{
		xenditClient: xenditClient,
		cache:        cache,
	}, nil
}

func (s *Service) CheckBalance(ctx context.Context) (float64, error) {
	return s.xenditClient.BalanceInquiry(ctx)
}

func (s *Service) GetInvoiceById(ctx context.Context, invoiceId string) (invoice.Invoice, error) {
	normalizedID := strings.TrimSpace(invoiceId)
	if normalizedID == "" {
		return invoice.Invoice{}, paymentdomain.ErrInvalidInput
	}
	return s.xenditClient.GetInvoiceById(ctx, normalizedID)
}

func (s *Service) CreateInvoice(ctx context.Context, input paymentdomain.CreateInvoiceInput) (invoice.Invoice, bool, error) {
	externalID := strings.TrimSpace(input.ExternalID)
	if externalID == "" || input.Amount <= 0 {
		return invoice.Invoice{}, false, paymentdomain.ErrInvalidInput
	}

	cacheKey := ""
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if idempotencyKey != "" && s.cache != nil {
		cacheKey = buildIdempotencyKey(idempotencyKey, externalID)
		cached, err := s.cache.GetString(ctx, cacheKey)
		if err == nil {
			if cached == idempotencyPendingValue {
				return invoice.Invoice{}, false, paymentdomain.ErrIdempotencyInProgress
			}
			cachedInvoice, err := s.xenditClient.GetInvoiceById(ctx, cached)
			if err != nil {
				return invoice.Invoice{}, false, err
			}
			return cachedInvoice, true, nil
		}
		if !errors.Is(err, redisinfra.ErrKeyNotFound) {
			return invoice.Invoice{}, false, err
		}

		acquired, err := s.cache.SetIfNotExists(ctx, cacheKey, idempotencyPendingValue, idempotencyPendingTTL)
		if err != nil {
			return invoice.Invoice{}, false, err
		}
		if !acquired {
			return invoice.Invoice{}, false, paymentdomain.ErrIdempotencyInProgress
		}
	}

	request := invoice.NewCreateInvoiceRequest(externalID, input.Amount)
	if value := normalizeOptionalString(input.PayerEmail); value != nil {
		request.PayerEmail = value
	}
	if value := normalizeOptionalString(input.Description); value != nil {
		request.Description = value
	}
	if input.InvoiceDurationSeconds != nil {
		if *input.InvoiceDurationSeconds <= 0 {
			return invoice.Invoice{}, false, paymentdomain.ErrInvalidInput
		}
		duration := float32(*input.InvoiceDurationSeconds)
		request.InvoiceDuration = &duration
	}
	if value := normalizeOptionalString(input.CallbackVirtualAccountID); value != nil {
		request.CallbackVirtualAccountId = value
	}
	if input.ShouldSendEmail != nil {
		request.ShouldSendEmail = input.ShouldSendEmail
	}
	if value := normalizeOptionalString(input.SuccessRedirectURL); value != nil {
		request.SuccessRedirectUrl = value
	}
	if value := normalizeOptionalString(input.FailureRedirectURL); value != nil {
		request.FailureRedirectUrl = value
	}
	if methods := normalizeStringSlice(input.PaymentMethods); len(methods) > 0 {
		request.PaymentMethods = methods
	}
	if value := normalizeOptionalString(input.MidLabel); value != nil {
		request.MidLabel = value
	}
	if input.ShouldAuthenticateCreditCard != nil {
		request.ShouldAuthenticateCreditCard = input.ShouldAuthenticateCreditCard
	}
	if value := normalizeOptionalString(input.Currency); value != nil {
		request.Currency = value
	}
	if input.ReminderTimeSeconds != nil {
		if *input.ReminderTimeSeconds <= 0 {
			return invoice.Invoice{}, false, paymentdomain.ErrInvalidInput
		}
		reminder := float32(*input.ReminderTimeSeconds)
		request.ReminderTime = &reminder
	}
	if value := normalizeOptionalString(input.Locale); value != nil {
		request.Locale = value
	}
	if value := normalizeOptionalString(input.ReminderTimeUnit); value != nil {
		request.ReminderTimeUnit = value
	}
	if len(input.Items) > 0 {
		items, err := mapInvoiceItems(input.Items)
		if err != nil {
			return invoice.Invoice{}, false, err
		}
		request.Items = items
	}
	if len(input.Fees) > 0 {
		fees, err := mapInvoiceFees(input.Fees)
		if err != nil {
			return invoice.Invoice{}, false, err
		}
		request.Fees = fees
	}
	if len(input.Metadata) > 0 {
		request.Metadata = input.Metadata
	}

	createdInvoice, err := s.xenditClient.CreateInvoice(ctx, *request)
	if err != nil {
		if cacheKey != "" && s.cache != nil {
			_ = s.cache.Delete(ctx, cacheKey)
		}
		return invoice.Invoice{}, false, err
	}

	if cacheKey != "" && s.cache != nil {
		if err := s.cache.SetWithTTL(ctx, cacheKey, createdInvoice.Id, idempotencyTTL); err != nil {
			logrus.WithFields(logrus.Fields{
				"cache_key":   cacheKey,
				"invoice_id":  createdInvoice.Id,
				"external_id": externalID,
			}).WithError(err).Warn("payment idempotency store failed")
		}
	}

	return createdInvoice, false, nil
}

func (s *Service) ExpireInvoice(ctx context.Context, invoiceId string) (invoice.Invoice, error) {
	normalizedID := strings.TrimSpace(invoiceId)
	if normalizedID == "" {
		return invoice.Invoice{}, paymentdomain.ErrInvalidInput
	}
	return s.xenditClient.ExpireInvoice(ctx, normalizedID)
}

func (s *Service) HandleInvoiceWebhook(ctx context.Context, payload invoice.InvoiceCallback, payloadHash string) error {
	if strings.TrimSpace(payload.Id) == "" || strings.TrimSpace(payload.ExternalId) == "" || strings.TrimSpace(payload.Status) == "" {
		return paymentdomain.ErrInvalidInput
	}
	if payload.Amount <= 0 {
		return paymentdomain.ErrInvalidInput
	}
	if strings.TrimSpace(payloadHash) == "" {
		return paymentdomain.ErrInvalidInput
	}

	if s.cache != nil {
		key := buildWebhookDedupKey(payload, payloadHash)
		acquired, err := s.cache.SetIfNotExists(ctx, key, time.Now().UTC().Format(time.RFC3339Nano), webhookDedupTTL)
		if err != nil {
			return err
		}
		if !acquired {
			logrus.WithFields(logrus.Fields{
				"invoice_id":  payload.Id,
				"external_id": payload.ExternalId,
				"status":      payload.Status,
			}).Info("xendit webhook duplicate ignored")
			return nil
		}
	}

	// TODO: persist payment status transition when payment storage is available.
	return nil
}

func buildIdempotencyKey(key, externalID string) string {
	trimmedKey := strings.TrimSpace(key)
	trimmedExternal := strings.TrimSpace(externalID)
	if trimmedExternal == "" {
		return "payment:invoice:idempotency:" + trimmedKey
	}
	return "payment:invoice:idempotency:" + trimmedKey + ":" + trimmedExternal
}

func buildWebhookDedupKey(payload invoice.InvoiceCallback, payloadHash string) string {
	invoiceID := strings.TrimSpace(payload.Id)
	status := strings.ToUpper(strings.TrimSpace(payload.Status))
	if invoiceID == "" {
		if status == "" {
			return "payment:webhook:xendit:invoice:hash:" + payloadHash
		}
		return "payment:webhook:xendit:invoice:status:" + status + ":hash:" + payloadHash
	}
	if status == "" {
		return "payment:webhook:xendit:invoice:" + invoiceID + ":hash:" + payloadHash
	}
	return "payment:webhook:xendit:invoice:" + invoiceID + ":status:" + status + ":hash:" + payloadHash
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func mapInvoiceItems(items []paymentdomain.InvoiceItem) ([]invoice.InvoiceItem, error) {
	result := make([]invoice.InvoiceItem, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" || item.Price < 0 || item.Quantity < 0 {
			return nil, paymentdomain.ErrInvalidInput
		}
		mapped := invoice.InvoiceItem{
			Name:     name,
			Price:    float32(item.Price),
			Quantity: float32(item.Quantity),
		}
		if value := normalizeOptionalString(item.ReferenceID); value != nil {
			mapped.ReferenceId = value
		}
		if value := normalizeOptionalString(item.URL); value != nil {
			mapped.Url = value
		}
		if value := normalizeOptionalString(item.Category); value != nil {
			mapped.Category = value
		}
		result = append(result, mapped)
	}
	return result, nil
}

func mapInvoiceFees(fees []paymentdomain.InvoiceFee) ([]invoice.InvoiceFee, error) {
	result := make([]invoice.InvoiceFee, 0, len(fees))
	for _, fee := range fees {
		feeType := strings.TrimSpace(fee.Type)
		if feeType == "" || fee.Value < 0 {
			return nil, paymentdomain.ErrInvalidInput
		}
		result = append(result, invoice.InvoiceFee{
			Type:  feeType,
			Value: float32(fee.Value),
		})
	}
	return result, nil
}
