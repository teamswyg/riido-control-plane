package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestAppendRetainedThreadProgressLinesKeepsLatestAppendOnlyLines(t *testing.T) {
	existing := make([]AgentThreadProgressLine, aiAgentClientThreadProgressLineLimit)
	for i := range existing {
		existing[i] = AgentThreadProgressLine{Seq: i + 1, Message: "existing"}
	}
	got := appendRetainedThreadProgressLines(existing, []AgentThreadProgressLine{
		{Seq: aiAgentClientThreadProgressLineLimit + 1, Message: "new"},
	})
	if len(got) != aiAgentClientThreadProgressLineLimit {
		t.Fatalf("line count = %d", len(got))
	}
	if got[0].Seq != 2 || got[len(got)-1].Seq != aiAgentClientThreadProgressLineLimit+1 {
		t.Fatalf("retained seq range = %d..%d", got[0].Seq, got[len(got)-1].Seq)
	}
}

func TestAppendRetainedThreadProgressLinesPreservesPartialReplacement(t *testing.T) {
	got := appendRetainedThreadProgressLines(
		[]AgentThreadProgressLine{
			{Seq: 1, Message: "thinking", MessageKey: progressmessage.AssistantPartialKey},
			{Seq: 2, Message: "stable"},
		},
		[]AgentThreadProgressLine{{
			Seq:        3,
			Message:    "still thinking",
			MessageKey: progressmessage.AssistantPartialKey,
		}},
	)
	if len(got) != 2 {
		t.Fatalf("lines = %+v", got)
	}
	if got[0].Seq != 3 || got[0].Message != "still thinking" {
		t.Fatalf("partial line was not replaced: %+v", got)
	}
	if got[1].Message != "stable" {
		t.Fatalf("stable line changed: %+v", got)
	}
}
