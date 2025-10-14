package main

import "net/http"

type HttpClientTransport struct{}

var httpClient = &http.Client{
	Transport: &HttpClientTransport{},
}

func (HttpClientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "thepool. (0.1, https://github.com/pmh-only/thepool)")
	return http.DefaultTransport.RoundTrip(req)
}
