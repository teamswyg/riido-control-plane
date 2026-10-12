package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPartialResponseDiscriminatorSurvivesWire(t *testing.T) {
	for _, key := range []string{"assistant.partial", "runtime.activity"} {
		b, err := json.Marshal(AgentThreadProgressLine{Seq: 2, Message: "hello", MessageKey: key})
		if err != nil {
			t.Fatal(err)
		}
		var got AgentThreadProgressLine
		if err = json.Unmarshal(b, &got); err != nil {
			t.Fatal(err)
		}
		if got.MessageKey != key {
			t.Fatalf("wire lost presentation: %s", b)
		}
	}
	b, _ := json.Marshal(AgentThreadProgressLine{MessageCode: 1101, MessageKey: "tool.collecting", MessageArgs: map[string]string{"label": "private"}})
	if strings.Contains(string(b), "message_args") {
		t.Fatalf("template args exposed: %s", b)
	}
}
func TestRuntimeFailureVisibleBeforeTerminal(t *testing.T) {
	for _, kind := range []string{EventProviderWarning, EventProviderError} {
		if got := assignmentEventVisibleThreadMessage(AssignmentRunning, kind, "quota exceeded", "working"); got != "quota exceeded" {
			t.Fatalf("hidden %s: %q", kind, got)
		}
	}
}

func TestPartialReplacementPreservesMonotonicSequence(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	s.taskThreads["sample"] = []AIAgentTaskThreadRecord{{TaskID: "sample", ThreadID: "thread", AssignmentID: "assignment", AssignmentState: AgentAssignmentStateRunning, Lines: []AgentThreadProgressLine{{Seq: 21, Message: "old", MessageKey: "assistant.partial"}}}}
	for i := 22; i <= 25; i++ {
		seq := s.nextThreadProgressSeqLocked("sample", "thread", nil)
		if seq != i {
			t.Fatalf("sequence reused after replacement: got %d want %d", seq, i)
		}
		s.appendThreadProgressLocked(AgentThreadProgressEvent{TaskID: "sample", ThreadID: "thread", AssignmentID: "assignment", AssignmentState: AgentAssignmentStateRunning, Lines: []AgentThreadProgressLine{{Seq: seq, Message: "new", MessageKey: "assistant.partial"}}})
	}
	if len(s.taskThreads["sample"][0].Lines) != 1 {
		t.Fatal("partial snapshots accumulated")
	}
}
