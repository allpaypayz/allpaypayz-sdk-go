package allpaypayz

// Wire-level types for the Allpaypayz API v4. Mirrors the schemas under
// allpaypays-core-api-doc/api-v4/schemas-en/* — kept hand-rolled so callers
// see field tags exactly matching the JSON shape.

// Money is the canonical amount type — integer minor units plus ISO currency.
type Money struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type Customer struct {
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Country string `json:"country,omitempty"`
}

type CustomerDocument struct {
	Type        string `json:"type"`
	Number      string `json:"number"`
	Citizenship string `json:"citizenship,omitempty"`
	BirthDate   string `json:"birth_date,omitempty"`
}

type Payee struct {
	Customer
	Document *CustomerDocument `json:"document,omitempty"`
}

type Card struct {
	Pan      string `json:"pan"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
	CVC      string `json:"cvc"`
	Holder   string `json:"holder,omitempty"`
}

type ThreeDSChallenge struct {
	AcsURL  string `json:"acs_url,omitempty"`
	CReq    string `json:"c_req,omitempty"`
	MD      string `json:"md,omitempty"`
	TermURL string `json:"term_url,omitempty"`
}

type ThreeDSDeviceData struct {
	AcceptHeader   string `json:"accept_header,omitempty"`
	ColorDepth     int    `json:"color_depth,omitempty"`
	JavaEnabled    bool   `json:"java_enabled,omitempty"`
	Language       string `json:"language,omitempty"`
	ScreenHeight   int    `json:"screen_height,omitempty"`
	ScreenWidth    int    `json:"screen_width,omitempty"`
	TimezoneOffset int    `json:"timezone_offset,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
}

type URLs struct {
	Success  string `json:"success,omitempty"`
	Error    string `json:"error,omitempty"`
	Back     string `json:"back,omitempty"`
	Callback string `json:"callback,omitempty"`
}

// ExtraData round-trips arbitrary merchant-supplied JSON on create / response /
// webhook payloads.
type ExtraData map[string]any

// ---------- Resources ----------

type Payment struct {
	ID                string            `json:"id"`
	MerchantReference string            `json:"merchant_reference"`
	Status            string            `json:"status"`
	StatusReason      *string           `json:"status_reason"`
	StatusMessage     *string           `json:"status_message,omitempty"`
	Amount            Money             `json:"amount"`
	Fee               *Money            `json:"fee,omitempty"`
	RefundedAmount    *Money            `json:"refunded_amount,omitempty"`
	CardBrand         *string           `json:"card_brand,omitempty"`
	CardLast4         *string           `json:"card_last4,omitempty"`
	ThreeDS           *ThreeDSChallenge `json:"three_ds,omitempty"`
	ExtraData         ExtraData         `json:"extra_data,omitempty"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

type Payout struct {
	ID                string         `json:"id"`
	MerchantReference string         `json:"merchant_reference"`
	Status            string         `json:"status"`
	StatusReason      *string        `json:"status_reason"`
	Amount            Money          `json:"amount"`
	Fee               *Money         `json:"fee,omitempty"`
	Destination       map[string]any `json:"destination,omitempty"`
	Payee             *Payee         `json:"payee,omitempty"`
	Method            string         `json:"method"`
	ExtraData         ExtraData      `json:"extra_data,omitempty"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
}

type P2PTransfer struct {
	ID                string         `json:"id"`
	MerchantReference string         `json:"merchant_reference"`
	Status            string         `json:"status"`
	StatusReason      *string        `json:"status_reason"`
	Amount            Money          `json:"amount"`
	Fee               *Money         `json:"fee,omitempty"`
	Method            string         `json:"method"`
	Receiver          map[string]any `json:"receiver,omitempty"`
	PayDeadlineAt     string         `json:"pay_deadline_at,omitempty"`
	ExtraData         ExtraData      `json:"extra_data,omitempty"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
}

type Order struct {
	ID                string    `json:"id"`
	MerchantReference string    `json:"merchant_reference"`
	Status            string    `json:"status"`
	Amount            Money     `json:"amount"`
	Payment           *Payment  `json:"payment,omitempty"`
	CheckoutURL       string    `json:"checkout_url,omitempty"`
	ExtraData         ExtraData `json:"extra_data,omitempty"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

type Refund struct {
	ID           string  `json:"id"`
	PaymentID    string  `json:"payment_id"`
	Status       string  `json:"status"`
	StatusReason *string `json:"status_reason"`
	Amount       Money   `json:"amount"`
	Reason       string  `json:"reason,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type Terminal struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Currency       string   `json:"currency"`
	PaymentMethods []string `json:"payment_methods"`
}

// ---------- Request shapes ----------

type CreatePaymentRequest struct {
	MerchantReference string             `json:"merchant_reference"`
	Amount            Money              `json:"amount"`
	Description       string             `json:"description,omitempty"`
	ClientIP          string             `json:"client_ip,omitempty"`
	Customer          *Customer          `json:"customer,omitempty"`
	Card              Card               `json:"card"`
	ThreeDS           *ThreeDSDeviceData `json:"three_ds,omitempty"`
	URLs              *URLs              `json:"urls,omitempty"`
	ExtraData         ExtraData          `json:"extra_data,omitempty"`
}

type CreatePaymentRedirectRequest struct {
	MerchantReference string    `json:"merchant_reference"`
	Amount            Money     `json:"amount"`
	Description       string    `json:"description,omitempty"`
	Customer          *Customer `json:"customer,omitempty"`
	PaymentMethod     string    `json:"payment_method,omitempty"`
	URLs              *URLs     `json:"urls,omitempty"`
	ExtraData         ExtraData `json:"extra_data,omitempty"`
}

type CreateRecurrentPaymentRequest struct {
	OriginalPaymentID string    `json:"original_payment_id"`
	Amount            Money     `json:"amount"`
	Description       string    `json:"description,omitempty"`
	ExtraData         ExtraData `json:"extra_data,omitempty"`
}

type Finish3DSRequest struct {
	CRes string `json:"c_res"`
	MD   string `json:"md,omitempty"`
}

type CreateRefundRequest struct {
	Amount      *Money `json:"amount,omitempty"`
	Description string `json:"description,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type CreatePayoutRequest struct {
	MerchantReference string         `json:"merchant_reference"`
	Amount            Money          `json:"amount"`
	Description       string         `json:"description,omitempty"`
	Method            string         `json:"method"`
	Destination       map[string]any `json:"destination"`
	Payee             Payee          `json:"payee"`
	URLs              *URLs          `json:"urls,omitempty"`
	ExtraData         ExtraData      `json:"extra_data,omitempty"`
}

type CreateP2PRequest struct {
	MerchantReference string    `json:"merchant_reference"`
	Amount            Money     `json:"amount"`
	Description       string    `json:"description,omitempty"`
	Method            string    `json:"method"`
	Customer          Customer  `json:"customer"`
	ClientIP          string    `json:"client_ip,omitempty"`
	URLs              *URLs     `json:"urls,omitempty"`
	ExtraData         ExtraData `json:"extra_data,omitempty"`
}

type ConfirmP2PRequest struct {
	Status       string `json:"status"`
	StatusReason string `json:"status_reason,omitempty"`
}

type CreateOrderRequest struct {
	MerchantReference string    `json:"merchant_reference"`
	Amount            Money     `json:"amount"`
	Description       string    `json:"description,omitempty"`
	Customer          *Customer `json:"customer,omitempty"`
	URLs              *URLs     `json:"urls,omitempty"`
	ExtraData         ExtraData `json:"extra_data,omitempty"`
}

// ---------- Envelopes ----------

type envelope[T any] struct {
	OK        bool   `json:"ok"`
	Data      T      `json:"data"`
	RequestID string `json:"request_id,omitempty"`
}

// ApiError describes the v4 error.* shape returned on every non-2xx response.
type ApiError struct {
	Type    string           `json:"type"`
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []map[string]any `json:"details,omitempty"`
}

type errorEnvelope struct {
	OK        bool     `json:"ok"`
	Error     ApiError `json:"error"`
	RequestID string   `json:"request_id,omitempty"`
}

// ---------- Webhooks ----------

type CallbackResource struct {
	Kind              string    `json:"kind"`
	ID                string    `json:"id"`
	MerchantReference string    `json:"merchant_reference,omitempty"`
	PaymentID         string    `json:"payment_id,omitempty"`
	Status            string    `json:"status"`
	StatusReason      *string   `json:"status_reason,omitempty"`
	Amount            Money     `json:"amount"`
	Fee               *Money    `json:"fee,omitempty"`
	RefundedAmount    *Money    `json:"refunded_amount,omitempty"`
	ExtraData         ExtraData `json:"extra_data,omitempty"`
	CreatedAt         string    `json:"created_at,omitempty"`
	UpdatedAt         string    `json:"updated_at,omitempty"`
}

type CallbackEvent struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	CreatedAt  string           `json:"created_at"`
	Version    string           `json:"version"`
	APIVersion string           `json:"api_version,omitempty"`
	Resource   CallbackResource `json:"resource"`
}

type CallbackEnvelope struct {
	OK    bool          `json:"ok"`
	Event CallbackEvent `json:"event"`
}
