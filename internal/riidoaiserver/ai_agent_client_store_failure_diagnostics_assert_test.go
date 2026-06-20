package riidoaiserver

import (
	"strings"
	"testing"
)

func assertFailureDiagnostics(t *testing.T, label string, diagnostics *AIAgentTaskThreadFailureDiagnostics) {
	t.Helper()
	if diagnostics == nil {
		t.Fatalf("%s failure diagnostics is nil", label)
	}
	if diagnostics.ResultStatus != "blocked" {
		t.Fatalf("%s result_status = %q", label, diagnostics.ResultStatus)
	}
	if diagnostics.FailureCategory != "provider_blocked" {
		t.Fatalf("%s failure_category = %q", label, diagnostics.FailureCategory)
	}
	if !strings.Contains(diagnostics.Message, "approval_timeout") {
		t.Fatalf("%s message = %q", label, diagnostics.Message)
	}
	for _, leaked := range []string{"/Users/", "/tmp/", "file://"} {
		if strings.Contains(diagnostics.Message, leaked) {
			t.Fatalf("%s diagnostics leaked local runtime path marker %q: %q", label, leaked, diagnostics.Message)
		}
	}
}
