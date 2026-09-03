package allpaypayz

import "context"

type TerminalResource struct {
	http *httpClient
}

func (t *TerminalResource) Get(ctx context.Context) (*Terminal, error) {
	var env envelope[Terminal]
	if err := t.http.do(ctx, "GET", "/v4/terminal", nil, nil, "", &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
