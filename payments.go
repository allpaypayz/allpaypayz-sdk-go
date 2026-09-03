package allpaypayz

import (
	"context"
	"net/url"
)

type Payments struct {
	http *httpClient
}

func (p *Payments) Create(ctx context.Context, req *CreatePaymentRequest, opts *RequestOpts) (*Payment, error) {
	return p.postPayment(ctx, "/v4/payments", req, opts)
}

func (p *Payments) CreateRedirect(ctx context.Context, req *CreatePaymentRedirectRequest, opts *RequestOpts) (*Payment, error) {
	return p.postPayment(ctx, "/v4/payments/redirect", req, opts)
}

func (p *Payments) Recurrent(ctx context.Context, req *CreateRecurrentPaymentRequest, opts *RequestOpts) (*Payment, error) {
	return p.postPayment(ctx, "/v4/payments/recurrent", req, opts)
}

func (p *Payments) Finish3DS(ctx context.Context, id string, req *Finish3DSRequest, opts *RequestOpts) (*Payment, error) {
	return p.postPayment(ctx, "/v4/payments/"+url.PathEscape(id)+"/finish-3ds", req, opts)
}

func (p *Payments) Get(ctx context.Context, id string) (*Payment, error) {
	var env envelope[Payment]
	if err := p.http.do(ctx, "GET", "/v4/payments/"+url.PathEscape(id), nil, nil, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *Payments) FindByReference(ctx context.Context, merchantReference string) (*Payment, error) {
	q := url.Values{}
	q.Set("merchant_reference", merchantReference)
	var env envelope[Payment]
	if err := p.http.do(ctx, "GET", "/v4/payments", nil, q, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *Payments) CreateRefund(ctx context.Context, paymentID string, req *CreateRefundRequest, opts *RequestOpts) (*Refund, error) {
	var env envelope[Refund]
	if err := p.http.do(ctx, "POST", "/v4/payments/"+url.PathEscape(paymentID)+"/refunds", req, nil, idempotencyKey(opts), &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *Payments) GetRefund(ctx context.Context, paymentID, refundID string) (*Refund, error) {
	path := "/v4/payments/" + url.PathEscape(paymentID) + "/refunds/" + url.PathEscape(refundID)
	var env envelope[Refund]
	if err := p.http.do(ctx, "GET", path, nil, nil, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *Payments) postPayment(ctx context.Context, path string, body any, opts *RequestOpts) (*Payment, error) {
	var env envelope[Payment]
	if err := p.http.do(ctx, "POST", path, body, nil, idempotencyKey(opts), &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func idempotencyKey(opts *RequestOpts) string {
	if opts == nil {
		return ""
	}
	return opts.IdempotencyKey
}
