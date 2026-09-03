package allpaypayz

import (
	"context"
	"net/url"
)

type Payouts struct {
	http *httpClient
}

func (p *Payouts) Create(ctx context.Context, req *CreatePayoutRequest, opts *RequestOpts) (*Payout, error) {
	var env envelope[Payout]
	if err := p.http.do(ctx, "POST", "/v4/payouts", req, nil, idempotencyKey(opts), &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *Payouts) Get(ctx context.Context, id string) (*Payout, error) {
	var env envelope[Payout]
	if err := p.http.do(ctx, "GET", "/v4/payouts/"+url.PathEscape(id), nil, nil, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *Payouts) FindByReference(ctx context.Context, merchantReference string) (*Payout, error) {
	q := url.Values{}
	q.Set("merchant_reference", merchantReference)
	var env envelope[Payout]
	if err := p.http.do(ctx, "GET", "/v4/payouts", nil, q, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
