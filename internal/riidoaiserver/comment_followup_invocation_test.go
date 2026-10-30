package riidoaiserver

import (
	"strings"
	"testing"
)

func TestCommentFollowupRestoresInvocationAndDispatches(t *testing.T) {
	f := newCommentCallFixtureWithContext(t, NewMemoryAssignmentContextStore())
	fresh := f.evidence.Context
	fresh.Document.Content = "Updated follow-up body: compare plans for twenty people."
	f.s.taskContext = &assignmentHTTPTaskContextReader{contextSnapshot: fresh}
	if w := f.complete(t, "Initial review done"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	principal := AuthorizationResult{PrincipalID: f.evidence.PrincipalID, WorkspaceID: f.evidence.WorkspaceID}
	req := CreateAIAgentTaskThreadMessageRequest{Body: "Review again and ask another Agent if useful", RequestID: "human-followup", SourceMessageID: "human-followup-message"}
	replies, err := f.s.prepareThreadReplies(t.Context(), principal, "", f.source.TaskID, f.source.ThreadID, req)
	if err != nil {
		t.Fatal(err)
	}
	next, err := f.s.publishThreadReply(t.Context(), principal, f.source.TaskID, f.source.ThreadID, replies[0])
	if err != nil {
		t.Fatal(err)
	}
	rows := f.threads(t)
	child, err := selectAIAgentThreadForFollowup(AIAgentTaskThreadCollectionResponse{Threads: rows}, next.ThreadID)
	if err != nil || child.CommentInvocation == nil {
		t.Fatalf("followup lost invocation: %+v %v", child, err)
	}
	projection, _, err := f.assignments.LoadAssignmentProjection(t.Context(), next.AssignmentID)
	if err != nil || !strings.Contains(projection.Assignment.Prompt, "policy_checked") {
		t.Fatalf("missing delegation guidance: %v", err)
	}
	reasons, err := f.s.precheckCommentMention(t.Context(), projection.Assignment, "agent-owned-codex")
	if err != nil || len(reasons) > 0 {
		t.Fatalf("followup precheck: %v %v", reasons, err)
	}
	// The binding and manual-chain marker survive persistent projection reload.
	f.client = openPersistentAIAgentClientStoreForDelegateTest(t, f.snapshots)
	f.s.aiAgent = f.client
	restored, _ := selectAIAgentThreadForFollowup(AIAgentTaskThreadCollectionResponse{Threads: f.threads(t)}, next.ThreadID)
	if restored.CommentInvocation == nil || !restored.CommentInvocation.HumanFollowup {
		t.Fatal("restart lost followup binding")
	}
	replay, err := f.s.prepareThreadReplies(t.Context(), principal, "", f.source.TaskID, f.source.ThreadID, req)
	if err != nil || replay[0].Prepared.Replay == nil || replay[0].Prepared.Replay.AssignmentID != next.AssignmentID {
		t.Fatal("retry created another call chain", err)
	}
	history, err := f.client.ListAIAgentTaskThreadHistory(t.Context(), principal, next.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	reacted := false
	for _, thread := range history.Threads {
		if thread.ThreadID == next.ThreadID {
			for _, msg := range thread.Messages {
				if msg.SourceMessageID == req.SourceMessageID || msg.MessageID == req.SourceMessageID {
					for _, reaction := range msg.Reactions {
						if reaction.Emoji == "1f440" {
							reacted = true
						}
					}
				}
			}
		}
	}
	if !reacted {
		t.Fatal("followup status reaction did not target the new human message")
	}
	f.source = next
	if w := f.complete(t, mentionHTML("agent-owned-codex")+"Review the same request"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	rows = f.threads(t)
	if len(rows) != 3 || rows[2].AgentID != "agent-owned-codex" || rows[2].ConversationID != "comment:root" {
		t.Fatalf("followup did not delegate: %+v", rows)
	}
	snapshot, found, err := f.s.config.AssignmentContext.LoadAssignmentContextSnapshot(t.Context(), rows[2].AssignmentID)
	if err != nil || !found || snapshot.Context.Document.Content != fresh.Document.Content {
		t.Fatalf("child lost fresh followup context: %+v %v", snapshot, err)
	}
}

func TestCommentFollowupDoesNotWriteStatusToOriginalComment(t *testing.T) {
	f := newCommentCallFixture(t)
	f.client.taskThreads[f.source.TaskID][0].CommentInvocation.HumanFollowup = true
	f.s.commentReactions = &commentReactionDispatcher{sent: make(map[string]string)}
	f.s.deliverCommentStatusReactions(t.Context(), AuthorizationResult{PrincipalID: f.evidence.PrincipalID, WorkspaceID: f.evidence.WorkspaceID}, f.source.TaskID)
	if f.s.commentReactions.alreadySent(f.evidence.EventID, "1f440") {
		t.Fatal("followup reaction was sent to original comment")
	}
}
