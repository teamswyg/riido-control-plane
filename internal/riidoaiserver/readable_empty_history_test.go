package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestAcceptanceRenderedEmptyHistoryPreservesEarlierRequirement(t *testing.T) {
	th := AIAgentTaskThreadRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent", ResultMessage: "previous final"}
	h := AIAgentTaskThreadHistoryRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent", Messages: []AIAgentTaskThreadHistoryMessage{{Role: AIAgentTaskThreadMessageRoleUser, Body: "Keep public API unchanged.", ObservedAt: time.Unix(1, 0)}}}
	for i := 0; i < 10; i++ {
		h.Messages = append(h.Messages, AIAgentTaskThreadHistoryMessage{Role: AIAgentTaskThreadMessageRoleUser, Body: "<p><br></p>", ObservedAt: time.Unix(int64(i+2), 0)})
	}
	got := appendAIAgentTaskThreadMessagePrompt("Task", th, CreateAIAgentTaskThreadMessageRequest{Body: "Continue."}, h)
	if !strings.Contains(got, "Keep public API unchanged.") {
		t.Fatal("Rendered-empty rich messages displaced the previous user requirement before history budgeting")
	}
}
