package riidoaiserver

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type runReadsFailStore struct{ *DevelopmentAIAgentClientStore }

func (*runReadsFailStore) FindAIAgentTaskThreadByRunID(context.Context, string, string) (AIAgentTaskThreadRecord, error) {
	return AIAgentTaskThreadRecord{}, errors.New("private persistence failure")
}

func TestAIAgentRunReadsStoreFailureAndScopeMismatch(t *testing.T) {
	server, _, _ := runReadsTestServer(t, []AssignmentContextRead{{SourceKind: "task", SourceID: "task", SourcePart: "body"}})
	base := server.aiAgent
	server.aiAgent = &runReadsFailStore{base.(*DevelopmentAIAgentClientStore)}
	response := runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads")
	if response.Code != 503 || strings.Contains(response.Body.String(), "private persistence") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	server.aiAgent = base
	store := server.config.AssignmentContext.(*MemoryAssignmentContextStore)
	store.reads["assignment"][0].TaskID = "another-task"
	response = runReadsRequest(server, "/v1/client/ai-agent/runs/run/reads")
	if response.Code != 503 || strings.Contains(response.Body.String(), "another-task") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAIAgentRunReadsCurrentVisibilityPartGranularity(t *testing.T) {
	current := AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task"}, Document: AIAgentTaskContextDocument{Status: "unavailable"}, Hierarchy: AIAgentTaskContextHierarchy{Project: AIAgentTaskContextReference{ID: "project", Document: &AIAgentTaskContextDocument{Content: "visible"}}}}
	visible := visibleRunReadSources(current)
	for _, key := range []runReadSourceKey{{"task", "task", "body"}, {"target", "private", "base"}, {"comment", "deleted", "body"}} {
		if visible[key] {
			t.Fatalf("unexpected grant %+v", key)
		}
	}
	for _, part := range []string{"title", "summary", "body"} {
		if !visible[runReadSourceKey{"project", "project", part}] {
			t.Fatalf("missing grant %s", part)
		}
	}
}
