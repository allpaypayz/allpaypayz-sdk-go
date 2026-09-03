package allpaypayz

import (
	"errors"
	"fmt"
)

// Error is the base SDK error. Every non-2xx response is mapped to one of the
// sentinel-tagged Error values below; callers compare with errors.Is or
// inspect Type / Code directly.
type Error struct {
	Type              string           // "validation", "conflict", "rate_limit", ..., "network"
	Code              string           // server-supplied code, e.g. "duplicate_reference"
	Message           string           // server-supplied message
	Status            int              // HTTP status, 0 for network-level errors
	RequestID         string           // server-supplied trace id
	Details           []map[string]any // structured per-field details, optional
	RetryAfterSeconds int              // populated for 429 responses
}

func (e *Error) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("allpaypayz: %s (%s, http %d)", e.Message, e.Code, e.Status)
	}
	return fmt.Sprintf("allpaypayz: %s (%s)", e.Message, e.Code)
}

// Sentinels — match with errors.Is. Each maps to one v4 error.type or the
// transport bucket.
var (
	ErrValidation     = errors.New("allpaypayz: validation")
	ErrAuthentication = errors.New("allpaypayz: authentication")
	ErrNotFound       = errors.New("allpaypayz: not_found")
	ErrConflict       = errors.New("allpaypayz: conflict")
	ErrBusiness       = errors.New("allpaypayz: business")
	ErrRateLimit      = errors.New("allpaypayz: rate_limit")
	ErrGateway        = errors.New("allpaypayz: gateway")
	ErrAPI            = errors.New("allpaypayz: api")
	ErrNetwork        = errors.New("allpaypayz: network")
)

func (e *Error) Is(target error) bool {
	switch target {
	case ErrValidation:
		return e.Type == "validation"
	case ErrAuthentication:
		return e.Type == "authentication"
	case ErrNotFound:
		return e.Type == "not_found"
	case ErrConflict:
		return e.Type == "conflict"
	case ErrBusiness:
		return e.Type == "business"
	case ErrRateLimit:
		return e.Type == "rate_limit"
	case ErrGateway:
		return e.Type == "gateway"
	case ErrAPI:
		return e.Type == "api"
	case ErrNetwork:
		return e.Type == "network"
	}
	return false
}

// WebhookError is raised by VerifyWebhook on any verification failure mode.
type WebhookError struct {
	Code    string // e.g. "signature_mismatch", "stale_delivery"
	Message string
}

func (e *WebhookError) Error() string {
	return fmt.Sprintf("allpaypayz.webhook: %s (%s)", e.Message, e.Code)
}

func statusToType(status int) string {
	switch {
	case status == 400:
		return "validation"
	case status == 401, status == 403:
		return "authentication"
	case status == 404:
		return "not_found"
	case status == 409:
		return "conflict"
	case status == 422:
		return "business"
	case status == 429:
		return "rate_limit"
	case status >= 500 && status <= 599:
		return "gateway"
	default:
		return "api"
	}
}

func buildAPIError(status int, env errorEnvelope, retryAfter int) *Error {
	errType := env.Error.Type
	if errType == "" {
		errType = statusToType(status)
	}
	code := env.Error.Code
	if code == "" {
		code = fmt.Sprintf("http_%d", status)
	}
	msg := env.Error.Message
	if msg == "" {
		msg = fmt.Sprintf("Request failed with status %d", status)
	}
	return &Error{
		Type:              errType,
		Code:              code,
		Message:           msg,
		Status:            status,
		RequestID:         env.RequestID,
		Details:           env.Error.Details,
		RetryAfterSeconds: retryAfter,
	}
}
