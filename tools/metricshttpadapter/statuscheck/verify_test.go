package statuscheck

import (
	"strings"
	"testing"
)

func TestVerifyAcceptsStatuses(t *testing.T) {
	required := []Required{
		{Case: "authorized", Status: 200},
		{Case: "missing_scope", Status: 403},
		{Case: "store_unconfigured", Status: 503},
	}
	result := Result{Authorized: 200, MissingScope: 403, Unconfigured: 503}
	if err := Verify(required, result); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyRejectsStatusMismatch(t *testing.T) {
	err := Verify([]Required{{Case: "authorized", Status: 200}}, Result{Authorized: 500})
	if err == nil || !strings.Contains(err.Error(), "status authorized = 500") {
		t.Fatalf("error = %v, want authorized mismatch", err)
	}
}
