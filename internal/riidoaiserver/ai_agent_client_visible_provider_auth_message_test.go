package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestClientVisibleTaskThreadTextLocalizesProviderAuthFailure(t *testing.T) {
	t.Parallel()
	cases := []string{
		"Failed to authenticate. API Error: 401 Invalid authentication credentials",
		"invalid authentication credentials",
		"API Error: 401",
	}
	for _, input := range cases {
		if got := clientVisibleTaskThreadText(input); got != clientMessageProviderAuthFailed {
			t.Fatalf("clientVisibleTaskThreadText(%q) = %q, want %q", input, got, clientMessageProviderAuthFailed)
		}
	}
}

func TestFailureDiagnosticsUsesProviderAuthCategory(t *testing.T) {
	t.Parallel()
	metadata := map[string]string{
		metadatakeys.AssignmentFailureCategory.String(): providerAuthFailureCategory,
	}
	diagnostics := failureDiagnosticsFromAssignmentEvent(metadata, "provider returned raw auth error")
	if diagnostics == nil {
		t.Fatal("expected diagnostics")
	}
	if diagnostics.Message != clientMessageProviderAuthFailed {
		t.Fatalf("diagnostics message = %q, want %q", diagnostics.Message, clientMessageProviderAuthFailed)
	}
}

func TestFailureDiagnosticsInfersProviderAuthFromRawMessage(t *testing.T) {
	t.Parallel()
	metadata := map[string]string{
		metadatakeys.AssignmentFailureCategory.String(): "provider_result_failed",
	}
	diagnostics := failureDiagnosticsFromAssignmentEvent(
		metadata,
		"Failed to authenticate. API Error: 401 Invalid authentication credentials",
	)
	if diagnostics == nil {
		t.Fatal("expected diagnostics")
	}
	if diagnostics.FailureCategory != providerAuthFailureCategory {
		t.Fatalf("failure_category = %q, want %q", diagnostics.FailureCategory, providerAuthFailureCategory)
	}
	if diagnostics.Message != clientMessageProviderAuthFailed {
		t.Fatalf("diagnostics message = %q, want %q", diagnostics.Message, clientMessageProviderAuthFailed)
	}
}

func TestAssignmentEventResponseLocalizesProviderAuthFailure(t *testing.T) {
	t.Parallel()
	response := assignmentEventActionResponse(
		AIAgentTaskThreadRecord{},
		AssignmentFailed,
		"Failed to authenticate. API Error: 401 Invalid authentication credentials",
		map[string]string{metadatakeys.AssignmentFailureCategory.String(): "provider_result_failed"},
	)
	if response.Message != clientMessageProviderAuthFailed ||
		response.ResultMessage != clientMessageProviderAuthFailed {
		t.Fatalf("response message/result = %q/%q", response.Message, response.ResultMessage)
	}
	if response.FailureDiagnostics == nil ||
		response.FailureDiagnostics.FailureCategory != providerAuthFailureCategory {
		t.Fatalf("failure diagnostics = %+v", response.FailureDiagnostics)
	}
}
