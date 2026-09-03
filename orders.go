package allpaypayz

import (
	"context"
	"net/url"
)

type Orders struct {
	http *httpClient
}

func (o *Orders) Create(ctx context.Context, req *CreateOrderRequest, opts *RequestOpts) (*Order, error) {
	var env envelope[Order]
	if err := o.http.do(ctx, "POST", "/v4/orders", req, nil, idempotencyKey(opts), &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (o *Orders) Get(ctx context.Context, id string) (*Order, error) {
	var env envelope[Order]
	if err := o.http.do(ctx, "GET", "/v4/orders/"+url.PathEscape(id), nil, nil, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (o *Orders) FindByReference(ctx context.Context, merchantReference string) (*Order, error) {
	q := url.Values{}
	q.Set("merchant_reference", merchantReference)
	var env envelope[Order]
	if err := o.http.do(ctx, "GET", "/v4/orders", nil, q, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
