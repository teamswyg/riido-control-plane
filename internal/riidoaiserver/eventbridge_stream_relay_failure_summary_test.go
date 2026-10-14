package riidoaiserver

import (
	"strings"
	"testing"
)

func TestEventBridgePutEventsResponseFailureSummary(t *testing.T) {
	if got := (eventBridgePutEventsResponse{FailedEntryCount: 3}).failureSummary(); got != "failed_entry_count=3" {
		t.Fatalf("empty summary = %q", got)
	}
	noCodes := eventBridgePutEventsResponse{
		FailedEntryCount: 1,
		Entries:          []eventBridgePutEventsResponseEntry{{EventID: "ok"}},
	}
	if got := noCodes.failureSummary(); got != "failed_entry_count=1" {
		t.Fatalf("no-code summary = %q", got)
	}
	withCodes := eventBridgePutEventsResponse{
		Entries: []eventBridgePutEventsResponseEntry{
			{ErrorCode: "Throttling", ErrorMessage: "slow"},
			{EventID: "ok"},
			{ErrorCode: "InternalFailure", ErrorMessage: "retry"},
		},
	}
	got := withCodes.failureSummary()
	if !strings.Contains(got, "Throttling") || !strings.Contains(got, "InternalFailure") {
		t.Fatalf("coded summary = %q", got)
	}
}
