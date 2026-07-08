package main

import (
	"strings"
	"testing"
)

func TestVerifyRoutingResultRejectsMismatch(t *testing.T) {
	t.Parallel()
	tc := routingCase{
		Name:              "case",
		WantAllowed:       true,
		WantRoutingStatus: "available",
		WantReason:        "provider available",
	}
	assertRoutingError(t, verifyRoutingResult(tc, routingEvidence{
		Name:    "case",
		Error:   "boom",
		Allowed: true,
	}), "got error")
	assertRoutingError(t, verifyRoutingResult(tc, routingEvidence{
		Name:          "case",
		Allowed:       true,
		RoutingStatus: "login-required",
		Reason:        "provider login required",
	}), "got login-required")
}

func TestVerifyRoutingResultRejectsMissingExpectedError(t *testing.T) {
	t.Parallel()
	err := verifyRoutingResult(
		routingCase{Name: "case", WantErrorContains: "runtime_provider"},
		routingEvidence{Name: "case", Error: "different"},
	)
	assertRoutingError(t, err, "want contains")
}

func assertRoutingError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want contains %q", err, want)
	}
}
