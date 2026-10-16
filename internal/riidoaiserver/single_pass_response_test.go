package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAnswerOnlyRequestsDoNotRequireIntentClarification(t *testing.T) {
	for _, title := range []string{"Summarize this planning note", "Explain the marketing plan", "Translate this document", "이 기획안을 요약", "이 차이를 설명"} {
		t.Run(title, func(t *testing.T) {
			got, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task", ComponentType: "task", Title: title}, Document: AIAgentTaskContextDocument{Content: "A planning note with enough context to answer."}}})
			if err != nil {
				t.Fatal(err)
			}
			if got.IntentGateRequired || !strings.Contains(got.Prompt, "execute_the_explicit_instruction") {
				t.Fatal("answer-only request was gated")
			}
			if !strings.Contains(got.Prompt, "answer directly") || !strings.Contains(got.Prompt, "do not start a separate") {
				t.Fatal("missing single-pass policy")
			}
		})
	}
}

func TestEmptyHistoryRetainsPreviousFinalAnswer(t *testing.T) {
	thread := AIAgentTaskThreadRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent", ResultMessage: "The agreed deadline is Friday."}
	record := AIAgentTaskThreadHistoryRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent", Messages: []AIAgentTaskThreadHistoryMessage{{Role: AIAgentTaskThreadMessageRoleAgent, Body: "  "}}}
	got := appendAIAgentTaskThreadMessagePrompt("Task: implement a change", thread, CreateAIAgentTaskThreadMessageRequest{Body: "Only explain the schedule; do not change files."}, record)
	if !strings.Contains(got, thread.ResultMessage) || !strings.Contains(got, "Only explain the schedule") {
		t.Fatal("lost latest directive or fallback result")
	}
	if strings.Contains(got, "Re-read the latest Task Document before answering") {
		t.Fatal("mandatory extra tool read restored")
	}
}

func TestEmptyRecentMessagesDoNotDisplaceUserRequirements(t *testing.T) {
	thread := AIAgentTaskThreadRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent"}
	record := AIAgentTaskThreadHistoryRecord{ThreadID: "thread", TaskID: "task", AgentID: "agent", Messages: []AIAgentTaskThreadHistoryMessage{{MessageID: "requirement", Role: AIAgentTaskThreadMessageRoleUser, Body: "Keep the public API unchanged."}}}
	for i := 0; i < 10; i++ {
		record.Messages = append(record.Messages, AIAgentTaskThreadHistoryMessage{Role: AIAgentTaskThreadMessageRoleAgent, Body: "  "})
	}
	got := appendAIAgentTaskThreadMessagePrompt("Task", thread, CreateAIAgentTaskThreadMessageRequest{Body: "Continue the requested implementation."}, record)
	if !strings.Contains(got, "Keep the public API unchanged.") {
		t.Fatal("empty status messages displaced the user requirement")
	}
}
