package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestCommentFollowupResetsOnlyHumanChainAndKeepsEvidenceIdentity(t *testing.T) {
	f := newCommentCallFixture(t)
	principal := AuthorizationResult{PrincipalID: f.evidence.PrincipalID, WorkspaceID: f.evidence.WorkspaceID}
	root := f.threads(t)[0]
	root.CommentInvocation.Depth = 3
	root.CommentInvocation.Remaining = 0
	root.CommentInvocation.Outcomes = map[string]string{"policy": "call_limit"}
	f.client.taskThreads[root.TaskID][0] = root
	next, err := f.s.prepareCommentFollowup(t.Context(), principal, root, "agent-owned-codex")
	if err != nil || next == nil {
		t.Fatalf("followup: %+v %v", next, err)
	}
	if next.Depth != 0 || next.Remaining != 8 || len(next.Ancestors) != 1 || next.Ancestors[0] != "agent-owned-codex" || len(next.Outcomes) != 0 || next.OriginMessageID != "" || !next.HumanFollowup {
		t.Fatalf("human chain not reset: %+v", next)
	}
	if !commentInvocationEvidenceMatches(f.evidence, root.TaskID, next) {
		t.Fatal("mention switch lost evidence identity")
	}
	if reasons := commentMentionLimitReasons(next, root.AgentID, 1); len(reasons) != 0 {
		t.Fatal("previous human chain leaked into new request", reasons)
	}
	if root.CommentInvocation.Depth != 3 || root.CommentInvocation.Remaining != 0 {
		t.Fatal("mutated original chain")
	}
	// Legacy records without a binding can recover only through their persisted parent.
	legacy := AIAgentTaskThreadRecord{ThreadID: "legacy-followup", AgentID: root.AgentID, TaskID: root.TaskID, ConversationID: root.ConversationID, ParentThreadID: root.ThreadID}
	f.client.taskThreads[root.TaskID] = append(f.client.taskThreads[root.TaskID], legacy)
	restored, err := f.s.prepareCommentFollowup(t.Context(), principal, legacy, root.AgentID)
	if err != nil || restored == nil || restored.EvidenceEventID != next.EvidenceEventID {
		t.Fatalf("legacy parent recovery: %+v %v", restored, err)
	}
}

func TestCommentFollowupCannotTransferEvidenceOrIgnoreRevocation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*commentCallFixture, *AIAgentTaskThreadRecord)
		wantErr bool
	}{
		{"different requester", func(f *commentCallFixture, r *AIAgentTaskThreadRecord) {
			r.CommentInvocation.PrincipalID = "another-user"
		}, false},
		{"different workspace", func(f *commentCallFixture, r *AIAgentTaskThreadRecord) {
			r.CommentInvocation.WorkspaceID = "another-workspace"
		}, false},
		{"different task", func(f *commentCallFixture, r *AIAgentTaskThreadRecord) {
			f.evidence.TaskID = "another-task"
			f.evidence.Context.Component.ID = "another-task"
		}, true},
		{"different root", func(f *commentCallFixture, r *AIAgentTaskThreadRecord) { f.evidence.RootCommentID = "another-root" }, true},
		{"access revoked", func(f *commentCallFixture, r *AIAgentTaskThreadRecord) { f.status = http.StatusForbidden }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCommentCallFixture(t)
			principal := AuthorizationResult{PrincipalID: f.evidence.PrincipalID, WorkspaceID: f.evidence.WorkspaceID}
			root := f.threads(t)[0]
			tc.mutate(f, &root)
			f.client.taskThreads[root.TaskID][0] = root
			next, err := f.s.prepareCommentFollowup(t.Context(), principal, root, root.AgentID)
			if next != nil || (err != nil) != tc.wantErr {
				t.Fatalf("unsafe followup: %+v %v", next, err)
			}
		})
	}
}
