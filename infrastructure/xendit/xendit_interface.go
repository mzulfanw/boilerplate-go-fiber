package xendit

import (
	"context"

	"github.com/xendit/xendit-go/v7/invoice"
)

type XenditClient interface {
	BalanceInquiry(ctx context.Context) (float64, error)
	GetInvoiceById(ctx context.Context, invoiceID string) (invoice.Invoice, error)
	CreateInvoice(ctx context.Context, req invoice.CreateInvoiceRequest) (invoice.Invoice, error)
	ExpireInvoice(ctx context.Context, invoiceID string) (invoice.Invoice, error)
}
