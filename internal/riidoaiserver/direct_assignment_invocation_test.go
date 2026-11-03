package riidoaiserver

import (
	"strings"
	"testing"
)

func TestDirectAssignmentFirstAnswerDelegatesOnceInOriginalConversation(t *testing.T) {
	for _, path := range []string{"assignment", "agent-assignments"} {
		t.Run(path, func(t *testing.T) {
			f := newDirectCallFixture(t, path, false)
			root := f.threads(t)[0]
			if root.CommentInvocation == nil || !root.CommentInvocation.matchesConversation(root) || root.SourceCommentID != "" {
				t.Fatalf("direct assignment lost root: %+v", root)
			}
			f.readHandoffCandidate(t, "agent-owned-codex")
			body := "<p>The program prints Hello, World! " + mentionHTML("agent-owned-codex") + " Please review the greeting format.</p>"
			if path == "agent-assignments" {
				body = "계산기 요구사항을 확인했습니다.\n\n테스트를 마쳤습니다.\n\n\n" + body
			}
			for range 2 {
				if w := f.complete(t, body); w.Code != 200 {
					t.Fatal(w.Body.String())
				}
			}
			rows := f.threads(t)
			if len(rows) != 2 {
				t.Fatalf("want two runs after replay; actual: %d, outcomes: %+v", len(rows), rows[0].CommentInvocation.Outcomes)
			}
			child := rows[1]
			if child.AgentID != "agent-owned-codex" || child.ThreadID == root.ThreadID || child.AssignmentID == root.AssignmentID || child.ParentThreadID != root.ThreadID || child.ConversationID != root.ThreadID || child.SourceCommentID != "" || child.SourceMessageID == "" || child.SourceMessageID != child.CommentInvocation.OriginMessageID {
				t.Fatalf("wrong reviewer, source or grouping: %+v", child)
			}
			ctx, found, err := f.s.config.AssignmentContext.LoadAssignmentContextSnapshot(t.Context(), child.AssignmentID)
			if err != nil || !found || ctx.Context.Document.Content != f.evidence.Context.Document.Content || ctx.TriggerID != child.SourceMessageID || ctx.RequesterPrincipalID != f.evidence.PrincipalID {
				t.Fatal("missing inherited review context", found, err)
			}
			proj, found, err := f.assignments.LoadAssignmentProjection(t.Context(), child.AssignmentID)
			if err != nil || !found || !strings.Contains(proj.Assignment.Prompt, "The program prints Hello, World!") {
				t.Fatal("child did not receive first answer", err)
			}
			f.source = actionResponseFromThread(child, f.evidence.WorkspaceID)
			if w := f.complete(t, "Reviewed: the greeting matches."); w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			history, err := f.client.ListAIAgentTaskThreadHistory(t.Context(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}, root.TaskID)
			if err != nil {
				t.Fatal(err)
			}
			authors := map[string]string{}
			for _, row := range history.Threads {
				for _, msg := range row.Messages {
					if msg.Role == "agent" && msg.MessageID != "" {
						authors[msg.MessageID] = row.AgentID
					}
				}
			}
			if len(authors) != 2 || authors[child.SourceMessageID] != root.AgentID {
				t.Fatalf("final message identities: %v", authors)
			}
		})
	}
}

func TestDirectAssignmentRechecksAccessBeforeReview(t *testing.T) {
	f := newDirectCallFixture(t, "agent-assignments", false)
	f.readHandoffCandidate(t, "agent-owned-codex")
	f.status = 403
	if w := f.complete(t, mentionHTML("agent-owned-codex")+"Please review."); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	rows := f.threads(t)
	if len(rows) != 1 || rows[0].CommentInvocation.Outcomes["policy"] != "access_revoked" {
		t.Fatalf("revoked user dispatched: %+v", rows)
	}
}
