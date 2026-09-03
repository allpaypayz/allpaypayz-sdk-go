package allpaypayz

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// RequestOpts carries the per-call knobs every resource method accepts as an
// optional second argument.
type RequestOpts struct {
	// IdempotencyKey overrides the auto-generated UUIDv4 the SDK injects on
	// every POST. Pass an empty string (or nil *RequestOpts) to use the
	// default; pass a value to deduplicate retries on the server.
	IdempotencyKey string
}

type httpClient struct {
	cfg config
}

func (h *httpClient) do(ctx context.Context, method, path string, body any, query url.Values, idempotencyKey string, out any) error {
	endpoint := strings.TrimRight(h.cfg.baseURL, "/") + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return &Error{Type: "api", Code: "request_serialize_failed", Message: err.Error()}
		}
		bodyBytes = b
	}

	var attemptErr error
	for attempt := 1; attempt <= h.cfg.retry.MaxAttempts; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
		if err != nil {
			return &Error{Type: "api", Code: "request_build_failed", Message: err.Error()}
		}
		req.Header.Set("Authorization", "Bearer "+h.cfg.apiKey)
		req.Header.Set("User-Agent", h.cfg.userAgent)
		req.Header.Set("Accept", "application/json")
		if h.cfg.apiVersion != "" {
			req.Header.Set("Accept-Api-Version", h.cfg.apiVersion)
		}
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if method == http.MethodPost {
			key := idempotencyKey
			if key == "" {
				key = newUUIDv4()
			}
			req.Header.Set("Idempotency-Key", key)
		}

		resp, err := h.cfg.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt < h.cfg.retry.MaxAttempts {
				attemptErr = err
				sleepBackoff(ctx, attempt, 0, h.cfg.retry)
				continue
			}
			return &Error{Type: "network", Code: "network_error", Message: err.Error()}
		}

		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 400 {
			if out == nil || len(data) == 0 {
				return nil
			}
			if err := json.Unmarshal(data, out); err != nil {
				return &Error{Type: "api", Code: "invalid_json_response", Message: err.Error(), Status: resp.StatusCode}
			}
			return nil
		}

		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		var env errorEnvelope
		_ = json.Unmarshal(data, &env)
		apiErr := buildAPIError(resp.StatusCode, env, retryAfter)

		if isRetryable(resp.StatusCode) && attempt < h.cfg.retry.MaxAttempts {
			attemptErr = apiErr
			sleepBackoff(ctx, attempt, retryAfter, h.cfg.retry)
			continue
		}
		return apiErr
	}
	if attemptErr != nil {
		return attemptErr
	}
	return &Error{Type: "api", Code: "retry_exhausted", Message: "all retries failed"}
}

func isRetryable(status int) bool {
	switch status {
	case 429, 500, 502, 503, 504:
		return true
	}
	return false
}

func parseRetryAfter(h string) int {
	if h == "" {
		return 0
	}
	if n, err := strconv.Atoi(strings.TrimSpace(h)); err == nil {
		return n
	}
	if t, err := time.Parse(time.RFC1123, h); err == nil {
		d := int(time.Until(t).Seconds())
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}

func sleepBackoff(ctx context.Context, attempt, retryAfter int, r RetryOptions) {
	var d time.Duration
	if retryAfter > 0 {
		d = time.Duration(retryAfter) * time.Second
	} else {
		exp := time.Duration(int64(r.InitialBackoff) << (attempt - 1))
		if exp > r.MaxBackoff {
			exp = r.MaxBackoff
		}
		jitter := time.Duration(mathrand.Int63n(int64(r.Jitter) + 1))
		d = exp + jitter
	}
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}

// newUUIDv4 produces a random UUIDv4 — uses crypto/rand for the bytes.
func newUUIDv4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
