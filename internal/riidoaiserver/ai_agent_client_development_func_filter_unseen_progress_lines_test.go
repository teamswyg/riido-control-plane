package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestFilterUnseenProgressLinesKeepsStrictlyNewLines(t *testing.T) {
	got := filterUnseenProgressLines(
		[]AgentThreadProgressLine{{Seq: 1}, {Seq: 2}},
		[]AgentThreadProgressLine{{Seq: 3}, {Seq: 4}},
	)
	if len(got) != 2 || got[0].Seq != 3 || got[1].Seq != 4 {
		t.Fatalf("new lines = %+v", got)
	}
}

func TestFilterUnseenProgressLinesDropsDuplicateIncomingSeq(t *testing.T) {
	got := filterUnseenProgressLines(
		[]AgentThreadProgressLine{{Seq: 1}},
		[]AgentThreadProgressLine{{Seq: 2}, {Seq: 2}},
	)
	if len(got) != 1 || got[0].Seq != 2 {
		t.Fatalf("duplicate seq was not filtered: %+v", got)
	}
}

func TestFilterUnseenProgressLinesKeepsReplacementLine(t *testing.T) {
	got := filterUnseenProgressLines(
		[]AgentThreadProgressLine{{Seq: 1}},
		[]AgentThreadProgressLine{{Seq: 1, MessageKey: progressmessage.AssistantPartialKey}},
	)
	if len(got) != 1 || got[0].MessageKey != progressmessage.AssistantPartialKey {
		t.Fatalf("replacement line was filtered: %+v", got)
	}
}

func TestFilterUnseenProgressLinesLargeBatchDropsSeenSeq(t *testing.T) {
	incoming := []AgentThreadProgressLine{
		{Seq: 1},
		{Seq: 2, MessageKey: progressmessage.AssistantPartialKey},
		{Seq: 3},
		{Seq: 3},
		{Seq: 4},
		{Seq: 5},
		{Seq: 6},
		{Seq: 7},
		{Seq: 8},
	}
	got := filterUnseenProgressLines(
		[]AgentThreadProgressLine{{Seq: 1}, {Seq: 4}},
		incoming,
	)
	wantSeqs := []int{2, 3, 5, 6, 7, 8}
	if len(got) != len(wantSeqs) {
		t.Fatalf("filtered lines = %+v", got)
	}
	for i, want := range wantSeqs {
		if got[i].Seq != want {
			t.Fatalf("line %d seq = %d, want %d; got %+v", i, got[i].Seq, want, got)
		}
	}
	if got[0].MessageKey != progressmessage.AssistantPartialKey {
		t.Fatalf("replacement line lost: %+v", got)
	}
}
