package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestMergeThreadProgressLinesAppendsWithoutDroppingExisting(t *testing.T) {
	got := mergeThreadProgressLines(
		[]AgentThreadProgressLine{{Seq: 1, Message: "first"}},
		[]AgentThreadProgressLine{{Seq: 2, Message: "second"}},
	)
	if len(got) != 2 {
		t.Fatalf("lines = %+v", got)
	}
	if got[0].Message != "first" || got[1].Message != "second" {
		t.Fatalf("unexpected merge order: %+v", got)
	}
}

func TestMergeThreadProgressLinesReplacesLatestPartial(t *testing.T) {
	got := mergeThreadProgressLines(
		[]AgentThreadProgressLine{
			{Seq: 1, Message: "thinking", MessageKey: progressmessage.AssistantPartialKey},
			{Seq: 2, Message: "stable"},
		},
		[]AgentThreadProgressLine{{Seq: 3, Message: "still thinking", MessageKey: progressmessage.AssistantPartialKey}},
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
