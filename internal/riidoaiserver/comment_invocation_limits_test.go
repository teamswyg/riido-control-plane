package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestCommentInvocationStopsAncestorAndDepthAfterChildCompletion(t *testing.T) {
	f := newCommentCallFixture(t)
	if w := f.complete(t, mentionHTML("agent-owned-codex")+`<p>Compare plans A and B by cost.</p>`); w.Code != 200 {
		t.Fatal(w.Code)
	}
	child := f.threads(t)[1]
	e := TaskEvent{Type: EventAssignmentCompleted, State: AssignmentCompleted, TaskID: child.TaskID, AssignmentID: child.AssignmentID, AgentID: child.AgentID, WorkspaceID: f.evidence.WorkspaceID, CreatedBy: f.evidence.PrincipalID, Message: mentionHTML(f.source.AgentID), At: time.Now()}
	if err := f.client.RecordAIAgentAssignmentEvent(t.Context(), child.AgentID, AgentEventRequest{}, e); err != nil {
		t.Fatal(err)
	}
	if err := f.s.dispatchCommentMentions(t.Context(), e); err != nil {
		t.Fatal(err)
	}
	rows := f.threads(t)
	if len(rows) != 2 || rows[1].CommentInvocation.Outcomes["target:"+f.source.AgentID] != "cycle" {
		t.Fatalf("ancestor cycle: %+v", rows)
	}
	for _, tc := range []struct{ depth, budget int }{{3, 8}, {0, 0}} {
		f := newCommentCallFixture(t)
		f.client.taskThreads[f.source.TaskID][0].CommentInvocation.Depth = tc.depth
		f.client.taskThreads[f.source.TaskID][0].CommentInvocation.Remaining = tc.budget
		if err := f.client.saveSnapshot(t.Context()); err != nil {
			t.Fatal(err)
		}
		if w := f.complete(t, mentionHTML("agent-owned-codex")); w.Code != 200 {
			t.Fatal(w.Code)
		}
		if len(f.threads(t)) != 1 || f.threads(t)[0].CommentInvocation.Outcomes["policy"] != "call_limit" {
			t.Fatal("limit bypassed")
		}
	}
}

func TestCommentInvocationScopeAndPartialFailure(t *testing.T) {
	f := newCommentCallFixture(t)
	body := `<p>` + mentionHTML("agent-private-cursor") + mentionHTML("agent-owned-codex") + ` compare A and B costs.</p>`
	if w := f.complete(t, body); w.Code != 200 {
		t.Fatalf("mixed result %d %s", w.Code, w.Body.String())
	}
	rows := f.threads(t)
	if len(rows) != 2 || rows[0].CommentInvocation.Outcomes["target:agent-private-cursor"] != "target_unavailable" || rows[1].CommentInvocation.Remaining != 3 {
		t.Fatalf("partial success not bounded: %+v", rows)
	}
	f = newCommentCallFixture(t)
	f.evidence.RootCommentID = "another-root"
	if w := f.complete(t, mentionHTML("agent-owned-codex")); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if len(f.threads(t)) != 1 || !strings.Contains(f.client.taskThreadMessages[f.source.ThreadID][1].Body, "연결") {
		t.Fatal("cross-root evidence accepted")
	}
}
