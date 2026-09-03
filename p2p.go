package allpaypayz

import (
	"context"
	"net/url"
)

type P2PTransfers struct {
	http *httpClient
}

func (p *P2PTransfers) Create(ctx context.Context, req *CreateP2PRequest, opts *RequestOpts) (*P2PTransfer, error) {
	var env envelope[P2PTransfer]
	if err := p.http.do(ctx, "POST", "/v4/p2p-transfers", req, nil, idempotencyKey(opts), &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *P2PTransfers) Confirm(ctx context.Context, id string, req *ConfirmP2PRequest, opts *RequestOpts) (*P2PTransfer, error) {
	var env envelope[P2PTransfer]
	path := "/v4/p2p-transfers/" + url.PathEscape(id) + "/confirm"
	if err := p.http.do(ctx, "POST", path, req, nil, idempotencyKey(opts), &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *P2PTransfers) Get(ctx context.Context, id string) (*P2PTransfer, error) {
	var env envelope[P2PTransfer]
	if err := p.http.do(ctx, "GET", "/v4/p2p-transfers/"+url.PathEscape(id), nil, nil, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (p *P2PTransfers) FindByReference(ctx context.Context, merchantReference string) (*P2PTransfer, error) {
	q := url.Values{}
	q.Set("merchant_reference", merchantReference)
	var env envelope[P2PTransfer]
	if err := p.http.do(ctx, "GET", "/v4/p2p-transfers", nil, q, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
