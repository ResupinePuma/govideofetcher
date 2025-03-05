package utils

import (
	"context"
	"io"
	"net/http"
)

func HTTPRequest(ctx context.Context, client http.Client, method string, url string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err = client.Do(req)
	if err != nil {
		return
	}
	return
}
