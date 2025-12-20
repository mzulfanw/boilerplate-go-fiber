package payment

type BalanceResponse struct {
	Balance float64 `json:"balance"`
}

type CreateInvoiceRequest struct {
	ExternalID                   string               `json:"external_id" validate:"required,notblank"`
	Amount                       float64              `json:"amount" validate:"required,gt=0"`
	PayerEmail                   *string              `json:"payer_email,omitempty" validate:"omitempty,email"`
	Description                  *string              `json:"description,omitempty" validate:"omitempty,notblank"`
	InvoiceDurationSeconds       *int                 `json:"invoice_duration,omitempty" validate:"omitempty,gt=0"`
	CallbackVirtualAccountID     *string              `json:"callback_virtual_account_id,omitempty" validate:"omitempty,notblank"`
	ShouldSendEmail              *bool                `json:"should_send_email,omitempty"`
	SuccessRedirectURL           *string              `json:"success_redirect_url,omitempty" validate:"omitempty,url"`
	FailureRedirectURL           *string              `json:"failure_redirect_url,omitempty" validate:"omitempty,url"`
	PaymentMethods               []string             `json:"payment_methods,omitempty" validate:"omitempty,dive,notblank"`
	MidLabel                     *string              `json:"mid_label,omitempty" validate:"omitempty,notblank"`
	ShouldAuthenticateCreditCard *bool                `json:"should_authenticate_credit_card,omitempty"`
	Currency                     *string              `json:"currency,omitempty" validate:"omitempty,notblank"`
	ReminderTimeSeconds          *int                 `json:"reminder_time,omitempty" validate:"omitempty,gt=0"`
	Locale                       *string              `json:"locale,omitempty" validate:"omitempty,notblank"`
	ReminderTimeUnit             *string              `json:"reminder_time_unit,omitempty" validate:"omitempty,notblank"`
	Items                        []InvoiceItemRequest `json:"items,omitempty" validate:"omitempty,dive"`
	Fees                         []InvoiceFeeRequest  `json:"fees,omitempty" validate:"omitempty,dive"`
	Metadata                     map[string]any       `json:"metadata,omitempty"`
}

type InvoiceItemRequest struct {
	Name        string  `json:"name" validate:"required,notblank"`
	Price       float64 `json:"price" validate:"required,gte=0"`
	Quantity    float64 `json:"quantity" validate:"required,gte=0"`
	ReferenceID *string `json:"reference_id,omitempty" validate:"omitempty,notblank"`
	URL         *string `json:"url,omitempty" validate:"omitempty,url"`
	Category    *string `json:"category,omitempty" validate:"omitempty,notblank"`
}

type InvoiceFeeRequest struct {
	Type  string  `json:"type" validate:"required,notblank"`
	Value float64 `json:"value" validate:"required,gte=0"`
}
