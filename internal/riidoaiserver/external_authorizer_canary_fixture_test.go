package riidoaiserver

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

const externalAuthCanary = "private-auth-canary-5824"

type externalAuthCanaryTransport struct {
	status int
	body   string
	err    error
}

func (f externalAuthCanaryTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &http.Response{StatusCode: f.status, Body: io.NopCloser(strings.NewReader(f.body)), Header: make(http.Header), Request: r}, nil
}

func externalAuthCanaryClient(t *testing.T, status int, body string, err error) *ExternalHTTPAuthorizer {
	t.Helper()
	a, e := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: "https://" + externalAuthCanary + ".invalid/check", APIKey: externalAuthCanary, HTTPClient: &http.Client{Transport: externalAuthCanaryTransport{status: status, body: body, err: err}}})
	if e != nil {
		t.Fatal("canary authorizer construction failed")
	}
	return a
}
