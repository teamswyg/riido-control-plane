package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestExternalAuthorizerTypedFailurePreservesCancellationDeadlineAndDenial(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		a := externalAuthCanaryClient(t, 0, "", cause)
		_, err := a.Authorize(context.Background(), externalAuthCanary, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead})
		var transport *url.Error
		if !errors.Is(err, cause) || !errors.As(err, &transport) {
			t.Fatal("typed diagnostics broke cause identity")
		}
		if externalAuthorizerRetryable(err) != (errors.Is(cause, context.DeadlineExceeded)) {
			t.Fatal("diagnostics changed retry policy")
		}
		want := "cancelled"
		if errors.Is(cause, context.DeadlineExceeded) {
			want = "timeout"
		}
		if externalAuthorizerOutcome(err) != want {
			t.Fatal("diagnostics changed cancellation or timeout outcome")
		}
	}
	for _, tc := range []struct {
		status   int
		sentinel error
	}{{http.StatusUnauthorized, ErrAuthorizationUnauthenticated}, {http.StatusForbidden, ErrAuthorizationForbidden}} {
		a := externalAuthCanaryClient(t, tc.status, externalAuthCanary, nil)
		_, err := a.Authorize(context.Background(), externalAuthCanary, AuthorizationRequest{Resource: AuthorizationResourceMetrics, Action: AuthorizationActionRead})
		if !errors.Is(err, tc.sentinel) || reflect.TypeOf(err) != reflect.TypeOf(tc.sentinel) || externalAuthorizerRetryable(err) {
			t.Fatal("diagnostics changed exact denied sentinel or retry policy")
		}
	}
}

func TestExternalAuthorizerDiagnosticAllowlistRejectsInjectedValues(t *testing.T) {
	for _, tc := range []struct {
		kind, detail         string
		status               int
		wantKind, wantDetail string
		wantStatus           int
	}{
		{externalAuthCanary, externalAuthCanary, 503, "unclassified", "none", 0},
		{"transport", externalAuthCanary, 503, "transport", "none", 0},
		{"invalid_response", externalAuthCanary, 999, "invalid_response", "none", 0},
		{"upstream_http", "schema", 503, "upstream_http", "none", 503},
	} {
		kind, detail, status := externalAuthorizerFailureFields(externalAuthorizerFailed(tc.kind, tc.detail, tc.status, errors.New(externalAuthCanary)))
		if kind != tc.wantKind || detail != tc.wantDetail || status != tc.wantStatus {
			t.Fatal("unbounded diagnostic left private error")
		}
	}
}
