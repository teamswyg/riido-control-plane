package riidoaiserver

import (
	"testing"
)

func TestCurrentReplyMetadataFreshnessRequiresLaterTimestamp(t *testing.T) {
	current := AIAgentTaskContext{Extended: &AIAgentTaskContextExtended{Comments: []AIAgentTaskContextComment{{ID: "reply", UpdatedAt: "2026-10-26T01:00:00Z", Truncated: true, MetadataOnly: true}}}}
	meta := currentRunReadMetadata(current)[runReadSourceKey{"comment", "reply", "body"}]
	for _, tc := range []struct {
		updated string
		changed bool
	}{{"2026-10-26T00:00:00Z", true}, {"2026-10-26T01:00:00Z", false}, {"2026-10-26T02:00:00Z", false}, {"", false}} {
		got := runReadModifiedSinceRead(AssignmentContextRead{Status: "served", SourceUpdatedAt: tc.updated, SourceVersion: assignmentContextVersion("")}, meta)
		if tc.changed {
			if got == nil || !*got {
				t.Fatal("later timestamp was not reported")
			}
		} else if got != nil {
			t.Fatal("metadata-only body comparison must remain unknown")
		}
	}
}
