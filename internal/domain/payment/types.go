package payment

type InvoiceItem struct {
	Name        string
	Price       float64
	Quantity    float64
	ReferenceID *string
	URL         *string
	Category    *string
}

type InvoiceFee struct {
	Type  string
	Value float64
}

type CreateInvoiceInput struct {
	IdempotencyKey               string
	ExternalID                   string
	Amount                       float64
	PayerEmail                   *string
	Description                  *string
	InvoiceDurationSeconds       *int
	CallbackVirtualAccountID     *string
	ShouldSendEmail              *bool
	SuccessRedirectURL           *string
	FailureRedirectURL           *string
	PaymentMethods               []string
	MidLabel                     *string
	ShouldAuthenticateCreditCard *bool
	Currency                     *string
	ReminderTimeSeconds          *int
	Locale                       *string
	ReminderTimeUnit             *string
	Items                        []InvoiceItem
	Fees                         []InvoiceFee
	Metadata                     map[string]any
}
