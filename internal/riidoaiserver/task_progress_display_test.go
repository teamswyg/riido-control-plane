package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestContextualProgressLiveAndColdProjection(t *testing.T) {
	for _, tc := range []struct {
		code                  int
		key, arg, description string
	}{
		{1101, "tool.collecting", "description", "기존 댓글과 답글 구조를 확인하고 있어요."},
		{1103, "tool.running", "description", "답글 전송 권한을 점검하고 있어요."},
		{1104, "tool.completed", "summary", "답글 전송 권한 검증을 마쳤어요."},
	} {
		t.Run(tc.key, func(t *testing.T) {
			line := AgentThreadProgressLine{Seq: 1, Message: "internal raw tool output", MessageCode: tc.code, MessageKey: tc.key, MessageArgs: map[string]string{tc.arg: tc.description, "label": "internal tool"}}
			rows := taskThreadProgressMessages(AIAgentTaskThreadRecord{ThreadID: "thread", Lines: []AgentThreadProgressLine{line}})
			for _, value := range []any{line, rows[0]} {
				body, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var wire map[string]any
				if err := json.Unmarshal(body, &wire); err != nil {
					t.Fatal(err)
				}
				if wire["display_message"] != tc.description {
					t.Fatalf("missing contextual progress: %s", body)
				}
				if strings.Contains(string(body), "message_args") {
					t.Fatalf("raw args exposed: %s", body)
				}
			}
			if line.MessageArgs[tc.arg] != tc.description {
				t.Fatal("source mutated")
			}
		})
	}
}

func TestContextualProgressRejectsUnclassifiedAndOversizedPayloads(t *testing.T) {
	for _, line := range []AgentThreadProgressLine{
		{Message: "short raw draft"},
		{MessageCode: 9001, MessageKey: "assistant.partial", MessageArgs: map[string]string{"description": "partial answer"}},
		{MessageCode: 1103, MessageKey: "runtime.activity", MessageArgs: map[string]string{"description": "wrong classification"}},
		{MessageCode: 1103, MessageKey: "tool.running", MessageArgs: map[string]string{"description": strings.Repeat("가", 181)}},
		{MessageCode: 1103, MessageKey: "tool.running", MessageArgs: map[string]string{"description": "one\ntwo"}},
	} {
		body, err := json.Marshal(line)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), "display_message") {
			t.Fatalf("unclassified content gained public description: %s", body)
		}
	}
}
