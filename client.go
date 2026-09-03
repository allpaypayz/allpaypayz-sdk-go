// Package allpaypayz is the official Allpaypayz API v4 SDK for Go.
//
// The Client type composes resource sub-clients (Payments, Payouts, P2P,
// Orders, Terminal). Every method takes a context.Context; cancellation /
// deadlines propagate to the underlying HTTP request.
package allpaypayz

import (
	"errors"
	"net/http"
	"time"
)

const (
	Version         = "0.1.0"
	defaultBaseURL  = "https://api4.allpaypayz.com"
	baseUserAgent   = "Allpaypayz-SDK-Go/" + Version
	defaultTimeout  = 30 * time.Second
	defaultMaxAttempts    = 3
	defaultInitialBackoff = 250 * time.Millisecond
	defaultMaxBackoff     = 4 * time.Second
	defaultJitter         = 250 * time.Millisecond
)

// RetryOptions controls the retry policy applied on 5xx, 429, and network
// errors. The default is exponential back-off with jitter capped at three
// attempts.
type RetryOptions struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Jitter         time.Duration
}

// Option mutates a Client during construction. Passed to NewClient as a
// variadic; see WithBaseURL etc.
type Option func(*config)

type config struct {
	apiKey      string
	baseURL     string
	apiVersion  string
	userAgent   string
	timeout     time.Duration
	retry       RetryOptions
	httpClient  *http.Client
}

// WithBaseURL overrides the default https://api4.allpaypayz.com — point this at
// staging or your own gateway in front of the API.
func WithBaseURL(v string) Option { return func(c *config) { c.baseURL = v } }

// WithAPIVersion sets the Accept-Api-Version header on every request.
func WithAPIVersion(v string) Option { return func(c *config) { c.apiVersion = v } }

// WithUserAgent appends a custom suffix to the default Allpaypayz-SDK-Go/<v>
// User-Agent.
func WithUserAgent(v string) Option { return func(c *config) { c.userAgent = baseUserAgent + " " + v } }

// WithRetry overrides the default retry policy.
func WithRetry(v RetryOptions) Option { return func(c *config) { c.retry = v } }

// WithHTTPClient injects a caller-owned http.Client (for proxies, custom
// transports, mTLS).
func WithHTTPClient(v *http.Client) Option { return func(c *config) { c.httpClient = v } }

// WithTimeout sets the per-request timeout when WithHTTPClient is not used.
func WithTimeout(v time.Duration) Option { return func(c *config) { c.timeout = v } }

// Client is the top-level Allpaypayz SDK client.
type Client struct {
	http     *httpClient
	Payments *Payments
	Payouts  *Payouts
	P2P      *P2PTransfers
	Orders   *Orders
	Terminal *TerminalResource
}

// NewClient constructs a Client. The apiKey is required; everything else has
// a sensible default.
//
//	client, err := allpaypayz.NewClient("sk_test_abc")
//	if err != nil { ... }
//	payment, err := client.Payments.Create(ctx, &allpaypayz.CreatePaymentRequest{...}, nil)
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("allpaypayz: api key is required")
	}
	cfg := config{
		apiKey:    apiKey,
		baseURL:   defaultBaseURL,
		userAgent: baseUserAgent,
		timeout:   defaultTimeout,
		retry: RetryOptions{
			MaxAttempts:    defaultMaxAttempts,
			InitialBackoff: defaultInitialBackoff,
			MaxBackoff:     defaultMaxBackoff,
			Jitter:         defaultJitter,
		},
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: cfg.timeout}
	}
	hc := &httpClient{cfg: cfg}
	c := &Client{http: hc}
	c.Payments = &Payments{http: hc}
	c.Payouts = &Payouts{http: hc}
	c.P2P = &P2PTransfers{http: hc}
	c.Orders = &Orders{http: hc}
	c.Terminal = &TerminalResource{http: hc}
	return c, nil
}
