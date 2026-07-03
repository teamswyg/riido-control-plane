package httpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/seedruntime"
)

func GetCatalogStatus(handler http.Handler) (int, error) {
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, authorizedRequest(http.MethodGet, "/v1/agent-catalog", ""))
	var out riidoaiserver.AgentCatalogListResponse
	return resp.Code, json.Unmarshal(resp.Body.Bytes(), &out)
}

func GetProviderStatus(handler http.Handler) (int, error) {
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, authorizedRequest(http.MethodGet, "/v1/agents/store-review-agent/provider-status", ""))
	var out riidoaiserver.ProviderStatusSyncResponse
	return resp.Code, json.Unmarshal(resp.Body.Bytes(), &out)
}

func PostPollStatus(handler http.Handler) int {
	body := `{"daemon_id":"review-demo-daemon","runtime_id":"review-demo-daemon:agent:store-review-agent:demo"}`
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, authorizedRequest(http.MethodPost, "/v1/agents/store-review-agent/poll", body))
	return resp.Code
}

func authorizedRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seedruntime.ReviewToken())
	return req
}
