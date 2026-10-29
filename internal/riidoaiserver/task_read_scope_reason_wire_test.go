package riidoaiserver

import "testing"

func TestTaskReadScopeAcceptsStructuredMentionProvenance(t *testing.T) {
	for _, kind := range []string{"mention", "client_snapshot_mention", "root_comment_mention", "trigger_comment_mention"} {
		t.Run(kind, func(t *testing.T) {
			if err := readScopeReasons([]taskReadScopeReason{{Kind: kind}}, AssignmentContentComponent, map[string]int{}); err != nil {
				t.Fatalf("producer mention provenance rejected: %v", err)
			}
			if err := readScopeReasons([]taskReadScopeReason{{Kind: kind}}, AssignmentContentMeeting, map[string]int{}); err == nil {
				t.Fatal("work mention provenance widened to meeting")
			}
		})
	}
	if err := readScopeReasons([]taskReadScopeReason{{Kind: "untrusted_future_mention"}}, AssignmentContentComponent, map[string]int{}); err == nil {
		t.Fatal("unknown provenance accepted")
	}
}
