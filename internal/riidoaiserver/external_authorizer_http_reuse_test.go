package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestExternalHTTPAuthorizerObservesReusedConnection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"schema_version":"riido-external-authorizer-response.v1","allowed":true,"principal_id":"synthetic"}`))
	}))
	defer srv.Close()
	a, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	recorder := &recordingTraceRecorder{}
	for range 2 {
		_, err = a.Authorize(WithTraceRecorder(context.Background(), recorder), "synthetic", AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, WorkspaceID: "synthetic"})
		if err != nil {
			t.Fatal(err)
		}
	}
	spans := recorder.snapshot()
	for i, span := range spans {
		attrs := span.Attributes
		if attrs["riido.external_authorizer.http_phase"] != "complete" {
			t.Fatal("successful request did not complete")
		}
		if attrs["riido.external_authorizer.connection_reused"] != strconv.Itoa(i) {
			t.Fatal("connection reuse not observed")
		}
		if attrs["riido.external_authorizer.first_byte_elapsed_ms"] == "-1" {
			t.Fatal("first byte not observed")
		}
	}
}
