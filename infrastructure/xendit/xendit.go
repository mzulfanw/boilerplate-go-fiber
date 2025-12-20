package xendit

import (
	"context"
	"errors"
	"strings"

	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	xendit "github.com/xendit/xendit-go/v7"
	"github.com/xendit/xendit-go/v7/invoice"
)

type Client struct {
	secretKey string
	publicKey string
}

var _ XenditClient = (*Client)(nil)

func New(cfg config.Config) (*Client, error) {
	return &Client{
		secretKey: cfg.XENDIT_SECRET_KEY,
		publicKey: cfg.XENDIT_PUBLIC_KEY,
	}, nil
}

func (c *Client) BalanceInquiry(ctx context.Context) (float64, error) {
	if c == nil {
		return 0, errors.New("xendit: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	client := xendit.NewClient(c.secretKey)
	balance, _, err := client.BalanceApi.GetBalance(ctx).Execute()
	if err != nil {
		return 0, err
	}
	if balance == nil {
		return 0, errors.New("xendit: empty balance response")
	}

	return float64(balance.Balance), nil
}

func (c *Client) GetInvoiceById(ctx context.Context, invoiceID string) (invoice.Invoice, error) {
	if c == nil {
		return invoice.Invoice{}, errors.New("xendit: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(invoiceID) == "" {
		return invoice.Invoice{}, errors.New("xendit: invoiceID is empty")
	}

	client := xendit.NewClient(c.secretKey)
	invoiceResp, _, err := client.InvoiceApi.GetInvoiceById(ctx, invoiceID).Execute()
	if err != nil {
		return invoice.Invoice{}, err
	}
	if invoiceResp == nil {
		return invoice.Invoice{}, errors.New("xendit: empty invoice response")
	}

	return *invoiceResp, nil
}

func (c *Client) CreateInvoice(ctx context.Context, req invoice.CreateInvoiceRequest) (invoice.Invoice, error) {
	if c == nil {
		return invoice.Invoice{}, errors.New("xendit: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(req.ExternalId) == "" {
		return invoice.Invoice{}, errors.New("xendit: external_id is empty")
	}
	if req.Amount <= 0 {
		return invoice.Invoice{}, errors.New("xendit: amount must be positive")
	}

	client := xendit.NewClient(c.secretKey)
	invoiceResp, _, err := client.InvoiceApi.CreateInvoice(ctx).
		CreateInvoiceRequest(req).
		Execute()
	if err != nil {
		return invoice.Invoice{}, err
	}
	if invoiceResp == nil {
		return invoice.Invoice{}, errors.New("xendit: empty invoice response")
	}

	return *invoiceResp, nil
}

func (c *Client) ExpireInvoice(ctx context.Context, invoiceID string) (invoice.Invoice, error) {
	if c == nil {
		return invoice.Invoice{}, errors.New("xendit: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(invoiceID) == "" {
		return invoice.Invoice{}, errors.New("xendit: invoiceID is empty")
	}

	client := xendit.NewClient(c.secretKey)
	invoiceResp, _, err := client.InvoiceApi.ExpireInvoice(ctx, invoiceID).Execute()
	if err != nil {
		return invoice.Invoice{}, err
	}
	if invoiceResp == nil {
		return invoice.Invoice{}, errors.New("xendit: empty invoice response")
	}

	return *invoiceResp, nil
}
