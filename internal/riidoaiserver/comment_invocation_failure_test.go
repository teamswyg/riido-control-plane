package riidoaiserver

import (
	"testing"
)

func TestCommentInvocationFailuresRemainInOriginalConversation(t *testing.T) {
	for _, tc := range []struct {
		name, target, want string
		status             int
	}{
		{"cycle", "agent-public-openclaw", "cycle", 200},
		{"private", "agent-private-cursor", "target_unavailable", 200},
		{"offline", "agent-owned-claude", "runtime_offline", 200},
		{"revoked", "agent-owned-codex", "access_revoked", 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCommentCallFixture(t)
			f.status = tc.status
			for range 2 {
				if w := f.complete(t, mentionHTML(tc.target)); w.Code != 200 {
					t.Fatalf("failure notice: %d %s", w.Code, w.Body.String())
				}
			}
			rows := f.threads(t)
			if len(rows) != 1 {
				t.Fatal("rejected call executed")
			}
			outcomes := rows[0].CommentInvocation.Outcomes
			if outcomes["target:"+tc.target] != tc.want && outcomes["policy"] != tc.want {
				t.Fatalf("wrong failure: %+v", outcomes)
			}
			messages := f.client.taskThreadMessages[f.source.ThreadID]
			count := 0
			for _, m := range messages {
				if m.Role == "system" {
					count++
					if m.Body == "" {
						t.Fatal("invisible failure")
					}
				}
			}
			if count != 1 {
				t.Fatalf("notice count %d", count)
			}
		})
	}
}
