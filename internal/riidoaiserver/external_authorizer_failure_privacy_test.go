package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"testing"
)

func TestExternalAuthorizerFailureDiagnosticsAreTypedBoundedAndPrivate(t *testing.T) {
	valid := `{"schema_version":"riido-external-authorizer-response.v1","allowed":true,"principal_id":"` + externalAuthCanary + `"}`
	for _, tc := range []struct {
		name, body, kind, detail string
		status                   int
		cause                    error
	}{
		{name: "transport", kind: "transport", detail: "none", cause: errors.New(externalAuthCanary)},
		{name: "upstream", body: externalAuthCanary, kind: "upstream_http", detail: "none", status: 503},
		{name: "decode", body: `{"` + externalAuthCanary + `":true}`, kind: "invalid_response", detail: "decode", status: 200},
		{name: "schema", body: `{"schema_version":"` + externalAuthCanary + `"}`, kind: "invalid_response", detail: "schema", status: 200},
		{name: "trailing", body: valid + ` {}`, kind: "invalid_response", detail: "decode", status: 200},
		{name: "principal", body: `{"schema_version":"riido-external-authorizer-response.v1","allowed":true}`, kind: "invalid_response", detail: "contract", status: 200},
		{name: "roles", body: `{"schema_version":"riido-external-authorizer-response.v1","allowed":true,"principal_id":"x","roles":["` + externalAuthCanary + `"]}`, kind: "invalid_response", detail: "contract", status: 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			original := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(original) })
			a := externalAuthCanaryClient(t, tc.status, tc.body, tc.cause)
			trace := &recordingTraceRecorder{}
			_, err := a.Authorize(WithTraceRecorder(context.Background(), trace), externalAuthCanary, AuthorizationRequest{Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionRead, WorkspaceID: externalAuthCanary})
			var failure externalAuthorizerFailure
			if !errors.As(err, &failure) {
				t.Fatal("upstream boundary lost typed diagnostic")
			}
			kind, detail, status := externalAuthorizerFailureFields(err)
			if kind != tc.kind || detail != tc.detail || status != tc.status {
				t.Fatal("bounded diagnostic does not match failure boundary")
			}
			spans := trace.snapshot()
			if len(spans) != 1 || spans[0].Attributes["riido.external_authorizer.failure_kind"] != tc.kind || spans[0].Attributes["riido.external_authorizer.response_error"] != tc.detail || spans[0].Attributes["riido.external_authorizer.upstream_http_status"] != strconv.Itoa(tc.status) {
				t.Fatal("trace lost diagnostic fields")
			}
			encoded, _ := json.Marshal(spans)
			if strings.Contains(logs.String()+string(encoded)+err.Error(), externalAuthCanary) {
				t.Fatal("private upstream material leaked through diagnostics")
			}
			for _, field := range []string{`failure_kind="` + tc.kind + `"`, `response_error="` + tc.detail + `"`, `upstream_http_status=`} {
				if !strings.Contains(logs.String(), field) {
					t.Fatal("failure log lost bounded diagnostic")
				}
			}
		})
	}
}
