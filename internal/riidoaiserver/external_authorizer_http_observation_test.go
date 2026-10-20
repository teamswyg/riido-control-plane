package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestExternalHTTPAuthorizerDistinguishesHeaderAndBodyDeadline(t *testing.T) {
	for _, phase := range []string{"response_headers", "response_body"} {
		t.Run(phase, func(t *testing.T) {
			var calls atomic.Int64
			release := make(chan struct{})
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if phase == "response_body" {
					w.WriteHeader(http.StatusOK)
					w.(http.Flusher).Flush()
				}
				<-release
			}))
			t.Cleanup(srv.Close)
			defer close(release)
			a, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: srv.URL, Timeout: 100 * time.Millisecond})
			if err != nil {
				t.Fatal(err)
			}
			recorder := &recordingTraceRecorder{}
			_, err = a.Authorize(WithTraceRecorder(context.Background(), recorder), "credential-canary", AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, WorkspaceID: "workspace-canary"})
			if !errors.Is(err, context.DeadlineExceeded) || calls.Load() != 1 {
				t.Fatalf("deadline/single attempt: %v, calls %d", err, calls.Load())
			}
			spans := recorder.snapshot()
			attrs := spans[0].Attributes
			if got := attrs["riido.external_authorizer.http_phase"]; got != phase {
				t.Fatalf("phase=%s, want %s", got, phase)
			}
			if attrs["riido.external_authorizer.dns_done_elapsed_ms"] != "-1" {
				t.Fatal("IP endpoint must not fabricate DNS timing")
			}
			for _, v := range attrs {
				if strings.Contains(v, "canary") || strings.Contains(v, srv.URL) {
					t.Fatal("private material in HTTP observations")
				}
			}
		})
	}
}
