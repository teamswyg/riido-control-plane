package riidoaiserver

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTaskReadScopeRetryRebuildsIdenticalRead(t *testing.T) {
	client, _, _ := taskScopeRetryFixture(t, http.StatusServiceUnavailable, 0, 0)
	var first string
	var calls atomic.Int32
	client.httpClient.Transport = taskScopeRetryTransport(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		identity := r.Method + r.URL.String() + r.Header.Get("Authorization") + string(body)
		if calls.Add(1) == 1 {
			first = identity
			return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
		}
		if identity != first {
			t.Error("retry changed query, variables or requester identity")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":{"getComponentReadScope":{"rootId":"task-private"}}}`))}, nil
	})
	var out taskReadScope
	err := client.readScopeGraphQL(t.Context(), "synthetic", readScopeQuery, "getComponentReadScope", map[string]string{"componentId": "task-private", "teamId": "team"}, &out)
	if err != nil || calls.Load() != 2 || out.RootID != "task-private" {
		t.Fatalf("calls=%d result=%+v err=%v", calls.Load(), out, err)
	}
}
