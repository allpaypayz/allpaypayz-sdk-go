package allpaypayz

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// VerifyOptions controls VerifyWebhook's tolerance window and clock. The zero
// value yields a 300 second tolerance against time.Now().
type VerifyOptions struct {
	ToleranceSeconds int
	Now              func() time.Time
}

var signatureRE = regexp.MustCompile(`^t=(\d+),v1=([0-9a-fA-F]+)$`)

// VerifyWebhook parses the Callback-Signature header, recomputes
// HMAC-SHA256(t + "." + raw_body, signKey), runs a constant-time compare via
// hmac.Equal, rejects deliveries outside the tolerance window, and finally
// JSON-unmarshals the envelope's event field. Returns *WebhookError on any
// failure mode with a machine-readable Code.
func VerifyWebhook(rawBody []byte, signatureHeader, signKey string, opts *VerifyOptions) (*CallbackEvent, error) {
	tolerance := 300
	now := time.Now
	if opts != nil {
		if opts.ToleranceSeconds > 0 {
			tolerance = opts.ToleranceSeconds
		}
		if opts.Now != nil {
			now = opts.Now
		}
	}

	m := signatureRE.FindStringSubmatch(signatureHeader)
	if m == nil {
		return nil, &WebhookError{
			Code:    "invalid_signature_header",
			Message: "Malformed Callback-Signature: " + signatureHeader,
		}
	}
	ts, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return nil, &WebhookError{Code: "invalid_signature_header", Message: err.Error()}
	}
	provided, err := hex.DecodeString(m[2])
	if err != nil {
		return nil, &WebhookError{Code: "invalid_signature_header", Message: err.Error()}
	}

	mac := hmac.New(sha256.New, []byte(signKey))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(rawBody)
	expected := mac.Sum(nil)
	if !hmac.Equal(provided, expected) {
		return nil, &WebhookError{Code: "signature_mismatch", Message: "Webhook signature does not match"}
	}

	currentUnix := now().Unix()
	if abs64(currentUnix-ts) > int64(tolerance) {
		return nil, &WebhookError{
			Code:    "stale_delivery",
			Message: fmt.Sprintf("Webhook timestamp %d outside %ds tolerance (now=%d)", ts, tolerance, currentUnix),
		}
	}

	if len(rawBody) == 0 {
		return nil, &WebhookError{Code: "invalid_envelope", Message: "Webhook body is empty"}
	}
	var env CallbackEnvelope
	if err := json.Unmarshal(rawBody, &env); err != nil {
		return nil, &WebhookError{Code: "invalid_json", Message: err.Error()}
	}
	if env.Event.Type == "" {
		return nil, &WebhookError{Code: "invalid_envelope", Message: "Webhook envelope missing event field"}
	}
	return &env.Event, nil
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// WebhookHandler is the function signature accepted by WebhookDispatcher.
type WebhookHandler func(*CallbackEvent) error

// WebhookDispatcher routes a verified CallbackEvent to a handler registered
// for its event.Type. Handlers may return an error; the dispatcher surfaces
// it to the caller so the HTTP layer can decide whether to ack or 5xx.
type WebhookDispatcher struct {
	handlers map[string]WebhookHandler
}

func NewWebhookDispatcher() *WebhookDispatcher {
	return &WebhookDispatcher{handlers: make(map[string]WebhookHandler)}
}

func (d *WebhookDispatcher) On(eventType string, h WebhookHandler) *WebhookDispatcher {
	d.handlers[eventType] = h
	return d
}

func (d *WebhookDispatcher) Dispatch(event *CallbackEvent) error {
	h, ok := d.handlers[event.Type]
	if !ok {
		return nil
	}
	return h(event)
}
