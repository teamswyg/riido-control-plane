package riidoaiserver

import (
	"strings"
	"testing"
)

func TestCommentCompletedHTTPDispatchesOnceAndRestoresAfterRestart(t *testing.T) {
	f := newCommentCallFixture(t)
	body := `<p>Find the lower-cost plan and validate the human acceptance criterion. ` + mentionHTML("agent-owned-codex") + `</p>`
	for range 2 {
		w := f.complete(t, body)
		if w.Code != 200 {
			t.Fatalf("completion: %d %s", w.Code, w.Body.String())
		}
	}
	rows := f.threads(t)
	if len(rows) != 2 {
		t.Fatalf("expected real child assignment, got %+v", rows)
	}
	child := rows[1]
	origin := taskThreadProjectionMessageID(rows[0])
	if child.SourceCommentID != f.evidence.CommentID || child.SourceMessageID != origin || child.CommentInvocation.OriginMessageID != origin || child.CommentInvocation.OriginThreadID != rows[0].ThreadID || child.CommentInvocation.OriginAssignmentID != f.source.AssignmentID {
		t.Fatalf("caller identity lost or confused with human evidence: %+v", child)
	}
	if child.ConversationID != "comment:root" || child.AgentID != "agent-owned-codex" || child.CommentInvocation.Depth != 1 || child.CommentInvocation.Remaining != 7 {
		t.Fatalf("lost handoff binding: %+v", child)
	}
	projection, found, err := f.assignments.LoadAssignmentProjection(t.Context(), child.AssignmentID)
	if err != nil || !found || !strings.Contains(projection.Assignment.Prompt, f.evidence.Body) || !strings.Contains(projection.Assignment.Prompt, "Find the lower-cost plan") || !strings.Contains(projection.Assignment.Prompt, "Keep the human acceptance criterion") || !strings.Contains(projection.Assignment.Prompt, "data-mention-id") {
		t.Fatalf("missing real assignment/context: %+v %v", projection, err)
	}
	if len(f.client.taskThreadMessages[child.ThreadID]) != 0 {
		t.Fatal("Agent request duplicated as a human reply")
	}
	f.client = openPersistentAIAgentClientStoreForDelegateTest(t, f.snapshots)
	f.s.aiAgent = f.client
	if w := f.complete(t, body); w.Code != 200 {
		t.Fatalf("restart retry: %d %s", w.Code, w.Body.String())
	}
	if len(f.threads(t)) != 2 {
		t.Fatal("restart duplicated execution")
	}
	if restored := f.threads(t)[1]; restored.SourceMessageID != origin || restored.CommentInvocation.OriginMessageID != origin {
		t.Fatalf("restart changed calling message: %+v", restored)
	}
}
