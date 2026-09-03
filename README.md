# `github.com/allpaypayz/allpaypayz-sdk-go` (Go)

**[⬇ Download the latest version](https://github.com/allpaypayz/allpaypayz-sdk-go/archive/refs/heads/main.zip)** · [Browse the code](https://github.com/allpaypayz/allpaypayz-sdk-go) · [MIT](LICENSE)

<sub>The archive is a snapshot of `main` — the current state of the SDK. Tagged releases will appear on the Releases page once the code leaves alpha.</sub>


Official Allpaypayz API v4 SDK for Go.

> Status: **alpha** (v0.1.0). Requires Go 1.21+.

## Install

```bash
go get github.com/allpaypayz/allpaypayz-sdk-go@v0.1.0
```

Zero third-party dependencies — pure stdlib (`net/http`, `encoding/json`,
`crypto/hmac`).

## Quick start

```go
package main

import (
    "context"
    "log"
    "os"

    allpaypayz "github.com/allpaypayz/allpaypayz-sdk-go"
)

func main() {
    client, err := allpaypayz.NewClient(os.Getenv("ALLPAYPAYZ_API_KEY"))
    if err != nil { log.Fatal(err) }

    p, err := client.Payments.Create(context.Background(), &allpaypayz.CreatePaymentRequest{
        MerchantReference: "ORDER-77",
        Amount:            allpaypayz.Money{AmountMinor: 1000, Currency: "USD"},
        Description:       "Order #77",
        Customer:          &allpaypayz.Customer{Name: "Jane Doe", Email: "jane@example.com"},
        Card: allpaypayz.Card{
            Pan: "4111111111111111", ExpMonth: 12, ExpYear: 2029,
            CVC: "123", Holder: "JANE DOE",
        },
    }, nil)
    if err != nil { log.Fatal(err) }

    log.Printf("created %s — status %s", p.ID, p.Status)
}
```

The SDK auto-injects an `Idempotency-Key` (UUIDv4 from `crypto/rand`) on every
POST. Override per call via `&allpaypayz.RequestOpts{IdempotencyKey: "your-key"}`.

## Configuration

```go
client, _ := allpaypayz.NewClient("sk_test_...",
    allpaypayz.WithBaseURL("https://staging-api4.allpaypayz.com"),
    allpaypayz.WithAPIVersion("2026-05-20"),
    allpaypayz.WithUserAgent("MyApp/2.0"),
    allpaypayz.WithRetry(allpaypayz.RetryOptions{
        MaxAttempts:    3,
        InitialBackoff: 250 * time.Millisecond,
        MaxBackoff:     4 * time.Second,
        Jitter:         250 * time.Millisecond,
    }),
    allpaypayz.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
)
```

Context propagation is built in — pass `context.WithTimeout` /
`context.WithCancel` and the SDK will honour the deadline.

## Resources

| Resource | Methods |
|---|---|
| `client.Payments` | `Create`, `CreateRedirect`, `Recurrent`, `Finish3DS`, `Get`, `FindByReference`, `CreateRefund`, `GetRefund` |
| `client.Payouts`  | `Create`, `Get`, `FindByReference` |
| `client.P2P`      | `Create`, `Confirm`, `Get`, `FindByReference` |
| `client.Orders`   | `Create`, `Get`, `FindByReference` |
| `client.Terminal` | `Get` |

## Errors

```go
import "errors"

if _, err := client.Payments.Create(ctx, req, nil); err != nil {
    if errors.Is(err, allpaypayz.ErrConflict) {
        var pe *allpaypayz.Error
        if errors.As(err, &pe) && pe.Code == "duplicate_reference" {
            // merchant_reference already used on this terminal
        }
    }
}
```

| HTTP / `error.type` | Sentinel |
|---|---|
| `400` / `validation` | `allpaypayz.ErrValidation` |
| `401`, `403` / `authentication` | `allpaypayz.ErrAuthentication` |
| `404` / `not_found` | `allpaypayz.ErrNotFound` |
| `409` / `conflict` | `allpaypayz.ErrConflict` |
| `422` / `business` | `allpaypayz.ErrBusiness` |
| `429` / `rate_limit` | `allpaypayz.ErrRateLimit` (`*Error.RetryAfterSeconds`) |
| `5xx` / `gateway` | `allpaypayz.ErrGateway` |
| Network / transport | `allpaypayz.ErrNetwork` |

Every error is a `*allpaypayz.Error` with fields `Type`, `Code`, `Status`,
`RequestID`, `Details`, `RetryAfterSeconds`. Sentinels work via
`errors.Is`; structured fields via `errors.As`.

## Webhooks

```go
http.HandleFunc("/webhooks/allpaypayz", func(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    event, err := allpaypayz.VerifyWebhook(
        body,
        r.Header.Get("Callback-Signature"),
        os.Getenv("ALLPAYPAYZ_SIGN_KEY"),
        nil,
    )
    if err != nil {
        var we *allpaypayz.WebhookError
        if errors.As(err, &we) {
            http.Error(w, we.Code, http.StatusBadRequest)
            return
        }
    }
    switch event.Type {
    case "payment.succeeded":
        markOrderPaid(event.Resource.MerchantReference)
    }
    w.WriteHeader(http.StatusOK)
})
```

`VerifyWebhook` parses `Callback-Signature` (`t=<unix>,v1=<hex>`), recomputes
`HMAC-SHA256(t + "." + raw_body, signKey)` via `crypto/hmac.Equal` for
constant-time comparison, and rejects deliveries outside the 300 s tolerance
window (overridable via `VerifyOptions`).

For type-discriminated dispatch:

```go
dispatcher := allpaypayz.NewWebhookDispatcher().
    On("payment.succeeded", func(e *allpaypayz.CallbackEvent) error { ... }).
    On("payment.failed",    func(e *allpaypayz.CallbackEvent) error { ... })
_ = dispatcher.Dispatch(event)
```

## Tests

```bash
go test ./...
```

`webhooks_test.go` loads `../spec/test-vectors.json` and checks
byte-identity with every other Allpaypayz SDK. `client_test.go` uses
`httptest.NewServer` as a real backend; no mocks needed.

## License

MIT
