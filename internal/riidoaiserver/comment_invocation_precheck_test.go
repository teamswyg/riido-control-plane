package riidoaiserver

import (
	"slices"
	"testing"
	"time"
)

func TestCommentMentionDefaultPolicyChecksRunningAssignmentWithoutPublishing(t *testing.T) {
	f := newCommentCallFixture(t)
	projection, found, err := f.assignments.LoadAssignmentProjection(t.Context(), f.source.AssignmentID)
	if err != nil || !found {
		t.Fatalf("source: %v %v", found, err)
	}
	if f.s.config.AgentMentionPrecheck == nil {
		t.Fatal("comment handoff policy not injected")
	}
	current := projection.Assignment
	for _, tc := range []struct {
		name, reason  string
		depth, budget int
		ancestor      string
	}{
		{name: "available", budget: 8},
		{name: "depth", depth: 3, budget: 8, reason: "depth_limit"},
		{name: "budget", budget: 0, reason: "budget_exhausted"},
		{name: "ancestor", budget: 8, ancestor: "agent-owned-codex", reason: "cycle"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			call := f.client.taskThreads[current.TaskID][0].CommentInvocation
			call.Depth, call.Remaining = tc.depth, tc.budget
			call.Ancestors = []string{current.AgentID}
			if tc.ancestor != "" {
				call.Ancestors = append(call.Ancestors, tc.ancestor)
			}
			if err := f.client.saveSnapshot(t.Context()); err != nil {
				t.Fatal(err)
			}
			data, _, ok, err := f.s.assignmentAgentDirectoryView(t.Context(), current, false, AgentAssignmentContextRequest{View: AssignmentContextViewAgent, Target: "agent-owned-codex"}, time.Now())
			detail, _ := data.(AgentDirectoryDetail)
			if err != nil || !ok || !detail.PolicyChecked {
				t.Fatalf("policy not checked: %+v %v", detail, err)
			}
			if detail.Agent.Callable != (tc.reason == "") || (tc.reason != "" && !slices.Contains(detail.Agent.UnavailableReasons, tc.reason)) {
				t.Fatalf("policy mismatch: %+v", detail)
			}
			if len(f.threads(t)) != 1 || call.Remaining != tc.budget {
				t.Fatal("directory mutated execution budget or published work")
			}
		})
	}
}

func TestCommentMentionPrecheckRejectsMismatchedAssignmentScope(t *testing.T) {
	for _, field := range []string{"task", "workspace", "principal", "agent"} {
		t.Run(field, func(t *testing.T) {
			f := newCommentCallFixture(t)
			projection, _, _ := f.assignments.LoadAssignmentProjection(t.Context(), f.source.AssignmentID)
			current := projection.Assignment
			switch field {
			case "task":
				current.TaskID = "other-task"
			case "workspace":
				current.WorkspaceID = "other-workspace"
			case "principal":
				current.CreatedBy = "other-user"
			case "agent":
				current.AgentID = "other-agent"
			}
			if f.s.config.AgentMentionPrecheck == nil {
				t.Fatal("comment handoff policy not injected")
			}
			reasons, err := f.s.config.AgentMentionPrecheck(t.Context(), current, "agent-owned-codex")
			if err == nil && len(reasons) == 0 {
				t.Fatal("mismatched assignment scope accepted")
			}
		})
	}
}

func TestCommentMentionPrecheckRejectsActiveTargetInSameConversation(t *testing.T) {
	f := newCommentCallFixture(t)
	if w := f.complete(t, mentionHTML("agent-owned-codex")); w.Code != 200 {
		t.Fatal(w.Code)
	}
	projection, _, _ := f.assignments.LoadAssignmentProjection(t.Context(), f.source.AssignmentID)
	if f.s.config.AgentMentionPrecheck == nil {
		t.Fatal("comment handoff policy not injected")
	}
	reasons, err := f.s.config.AgentMentionPrecheck(t.Context(), projection.Assignment, "agent-owned-codex")
	if err != nil || !slices.Contains(reasons, "session_conflict") {
		t.Fatalf("active target: %v %v", reasons, err)
	}
	if len(f.threads(t)) != 2 {
		t.Fatal("precheck started another run")
	}
}
