package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAgentDirectoryAppliesMentionPrecheckWhenConfigured(t *testing.T) {
	ctx := context.Background()
	current := Assignment{ID: "asg-a", TaskID: "task-private", AgentID: "agent-owned-codex", CreatedBy: "user-1"}
	var asked []string
	precheck := func(_ context.Context, run Assignment, target string) ([]string, error) {
		if run.ID != current.ID {
			t.Fatalf("precheck got run %q", run.ID)
		}
		asked = append(asked, target)
		if target == "agent-owned-claude" {
			return []string{"depth_limit"}, nil
		}
		return nil, nil
	}
	server := NewServer(ServerConfig{AIAgentClient: NewDevelopmentAIAgentClientStore(), AgentMentionPrecheck: precheck})
	data, _, ok, err := server.assignmentAgentDirectoryView(ctx, current, false, AgentAssignmentContextRequest{View: AssignmentContextViewAgents}, time.Now())
	page, _ := data.(AgentDirectoryPage)
	if err != nil || !ok || !page.PolicyChecked {
		t.Fatalf("page=%+v ok=%v err=%v", page, ok, err)
	}
	for _, entry := range page.Agents {
		if entry.AgentID == "agent-owned-claude" && (entry.Callable || entry.UnavailableReasons[len(entry.UnavailableReasons)-1] != "depth_limit") {
			t.Fatalf("policy reason not applied: %+v", entry)
		}
	}
	for _, target := range asked {
		if target == current.AgentID {
			t.Fatal("the current agent is never prechecked")
		}
	}
	failing := NewServer(ServerConfig{AIAgentClient: NewDevelopmentAIAgentClientStore(), AgentMentionPrecheck: func(context.Context, Assignment, string) ([]string, error) {
		return nil, errors.New("policy store down")
	}})
	if _, _, _, err := failing.assignmentAgentDirectoryView(ctx, current, false, AgentAssignmentContextRequest{View: AssignmentContextViewAgent, Target: "agent-owned-claude"}, time.Now()); err == nil {
		t.Fatal("a failing precheck must not be shown as checked")
	}
	plain := NewServer(ServerConfig{AIAgentClient: NewDevelopmentAIAgentClientStore()})
	data, _, _, _ = plain.assignmentAgentDirectoryView(ctx, current, false, AgentAssignmentContextRequest{View: AssignmentContextViewAgent, Target: "agent-owned-claude"}, time.Now())
	if detail, _ := data.(AgentDirectoryDetail); detail.PolicyChecked || detail.Agent.AgentID != "agent-owned-claude" {
		t.Fatalf("unconfigured precheck: %+v", detail)
	}
}
