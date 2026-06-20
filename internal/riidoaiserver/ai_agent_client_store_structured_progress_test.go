package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDevelopmentAIAgentClientStoreRendersStructuredProgressCode(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	response, err := store.RecordAIAgentThreadProgress(context.Background(), "agent-owned-codex", AgentThreadProgressBatchRequest{
		AssignmentID: "asn-1",
		TaskID:       "task-1",
		RunID:        "run-1",
		Lines: []AgentThreadProgressLine{{
			Seq:         1,
			MessageCode: 1101,
			MessageArgs: map[string]string{
				"label":       "팀 프로젝트",
				"description": "팀의 프로젝트 목록, 진행 상태, 우선순위와 담당자 정보를 조회해 요약을 준비 중. . .",
			},
		}},
	})
	if err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	line := response.Event.Lines[0]
	want := "팀 프로젝트 수집 중 - 팀의 프로젝트 목록, 진행 상태, 우선순위와 담당자 정보를 조회해 요약을 준비 중. . ."
	if line.Message != want || line.MessageCode != 1101 || line.MessageKey != "tool.collecting" {
		t.Fatalf("line = %+v, want rendered structured progress", line)
	}
	body, err := json.Marshal(response.Event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	if strings.Contains(string(body), "message_code") || strings.Contains(string(body), "message_args") {
		t.Fatalf("structured progress internals leaked to client JSON: %s", string(body))
	}
	threads, err := store.ListAIAgentTaskThreads(context.Background(), AuthorizationResult{PrincipalID: "user-1"}, "task-1")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	for _, thread := range threads.Threads {
		for _, line := range thread.Lines {
			if line.Message == want {
				return
			}
		}
	}
	t.Fatalf("threads = %+v", threads)
}
