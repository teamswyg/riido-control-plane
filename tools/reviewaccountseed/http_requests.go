package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func authorizedRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+reviewToken())
	return req
}

func getCatalogStatus(handler http.Handler) (int, error) {
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, authorizedRequest(http.MethodGet, "/v1/agent-catalog", ""))
	var out riidoaiserver.AgentCatalogListResponse
	return resp.Code, json.Unmarshal(resp.Body.Bytes(), &out)
}

func getProviderStatus(handler http.Handler) (int, error) {
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, authorizedRequest(http.MethodGet, "/v1/agents/store-review-agent/provider-status", ""))
	var out riidoaiserver.ProviderStatusSyncResponse
	return resp.Code, json.Unmarshal(resp.Body.Bytes(), &out)
}

func postPollStatus(handler http.Handler) int {
	body := `{"daemon_id":"review-demo-daemon","runtime_id":"review-demo-daemon:agent:store-review-agent:demo"}`
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, authorizedRequest(http.MethodPost, "/v1/agents/store-review-agent/poll", body))
	return resp.Code
}
