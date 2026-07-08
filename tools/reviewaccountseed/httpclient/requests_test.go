package httpclient

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestStatusRequestsUseAuthorizedReviewToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer review-token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/v1/agent-catalog":
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"schema_version":"test","agents":[]}`))
		case "/v1/agents/store-review-agent/provider-status":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"schema_version":"test"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	if code, err := GetCatalogStatus(handler); code != http.StatusAccepted || err != nil {
		t.Fatalf("catalog code/error = %d/%v", code, err)
	}
	if code, err := GetProviderStatus(handler); code != http.StatusCreated || err != nil {
		t.Fatalf("provider code/error = %d/%v", code, err)
	}
}

func TestPostPollStatusUsesDaemonBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != http.MethodPost || !strings.Contains(string(body), "review-demo-daemon") {
			t.Fatalf("request = %s %s body=%s", r.Method, r.URL.Path, body)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if got := PostPollStatus(handler); got != http.StatusNoContent {
		t.Fatalf("status = %d", got)
	}
}
