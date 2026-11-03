package riidoaiserver

import (
	"strings"
	"testing"
)

func TestDirectAssignmentLegacyFollowupBindsOnlyNewRequest(t *testing.T) {
	f := newDirectCallFixture(t, "agent-assignments", true)
	if w := f.complete(t, mentionHTML("agent-owned-codex")+"Old review."); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	rows := f.threads(t)
	if len(rows) != 1 || rows[0].CommentInvocation != nil {
		t.Fatal("historical answer was replayed")
	}
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	other := p
	other.PrincipalID = "another"
	if call, err := f.s.directAssignmentFollowup(t.Context(), other, rows[0], rows[0].AgentID); err != nil || call != nil {
		t.Fatal("another requester recovered delegation", err)
	}
	req := CreateAIAgentTaskThreadMessageRequest{Body: "Please ask a reviewer to verify the greeting.", RequestID: "new-human-request", SourceMessageID: "new-human-message"}
	replies, err := f.s.prepareThreadReplies(t.Context(), p, "", f.source.TaskID, f.source.ThreadID, req)
	if err != nil {
		t.Fatal(err)
	}
	next, err := f.s.publishThreadReply(t.Context(), p, f.source.TaskID, f.source.ThreadID, replies[0])
	if err != nil {
		t.Fatal(err)
	}
	f.source = next
	rows = f.threads(t)
	if len(rows) != 2 || rows[1].CommentInvocation == nil || !rows[1].CommentInvocation.HumanFollowup || rows[1].CommentInvocation.DirectRoot.ThreadID != rows[0].ThreadID {
		t.Fatal("followup did not bind original direct assignment")
	}
	if w := f.complete(t, "<p>"+mentionHTML("agent-owned-codex")+" Please verify Hello, World!</p>"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	rows = f.threads(t)
	if len(rows) != 3 || rows[2].ConversationID != rows[0].ThreadID || rows[2].ParentThreadID != next.ThreadID {
		t.Fatalf("review did not follow new message: %+v", rows)
	}
}

func TestDirectAssignmentPlainResponseAndExamplesDoNotDelegate(t *testing.T) {
	f := newDirectCallFixture(t, "agent-assignments", false)
	if w := f.complete(t, "<p>Finished without review.</p><pre>"+mentionHTML("agent-owned-codex")+"</pre>"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := f.complete(t, mentionHTML("agent-owned-codex")); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if rows := f.threads(t); len(rows) != 1 || !strings.Contains(rows[0].ResultMessage, "Finished without review") {
		t.Fatal("example or changed terminal replay caused delegation")
	}
}

func TestDirectAssignmentRejectsBrokenLegacyParentChain(t *testing.T) {
	root := AIAgentTaskThreadRecord{TaskID: "task1", ThreadID: "root", ConversationID: "root"}
	child := AIAgentTaskThreadRecord{TaskID: "task1", ThreadID: "child", ConversationID: "root", ParentThreadID: "root"}
	for _, alter := range []func(*AIAgentTaskThreadRecord, *AIAgentTaskThreadRecord){func(_, c *AIAgentTaskThreadRecord) { c.ParentThreadID = "missing" }, func(_, c *AIAgentTaskThreadRecord) { c.ParentThreadID = c.ThreadID }, func(r, _ *AIAgentTaskThreadRecord) { r.TaskID = "other" }, func(r, _ *AIAgentTaskThreadRecord) { r.SourceCommentID = "human" }, func(r, _ *AIAgentTaskThreadRecord) { r.ConversationID = "different" }} {
		r, c := root, child
		alter(&r, &c)
		if _, ok := verifiedDirectConversationRoot(c, []AIAgentTaskThreadRecord{r, c}); ok {
			t.Fatal("accepted unrelated root", r, c)
		}
	}
}
