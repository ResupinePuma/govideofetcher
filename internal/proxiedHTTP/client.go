package proxiedHTTP

import (
	"net/http"
	"net/url"
)

func NewProxiedHTTPClient(proxyURL string) http.Client {
	u, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(u),
	}

	c := http.Client{
		Transport: transport,
	}

	return c
}
