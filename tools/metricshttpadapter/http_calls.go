package main

import (
	"net/http"
	"net/http/httptest"
)

func callMetrics(handler http.Handler, m manifest, token string) (int, []byte) {
	req := httptest.NewRequest(m.Endpoint.Method, m.Endpoint.Path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp.Code, resp.Body.Bytes()
}
