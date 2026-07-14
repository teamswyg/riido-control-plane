package riidoaiserver

import (
	"net/http"
	"time"
)

func externalAuthorizerHTTPClient(client *http.Client, timeout time.Duration) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: timeout}
}
