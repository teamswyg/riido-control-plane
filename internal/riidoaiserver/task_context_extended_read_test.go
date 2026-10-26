package riidoaiserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtendedTaskContextReadsTitlesAndPrioritizedComments(t *testing.T) {
	var commentPages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/documents/providers/team/task":
			if r.URL.Query().Get("format") == "json" {
				writeTaskContextFixture(t, w, map[string]any{"tiptapJson": nil})
				return
			}
			writeTaskContextFixture(t, w, map[string]any{"html": "<p>body</p>"})
		case "/teams/team/components/task/files":
			writeTaskContextFixture(t, w, []map[string]any{})
		case "/teams/team/components/task/web-links":
			writeTaskContextFixture(t, w, map[string]any{"links": []map[string]any{}})
		case "/public/components/task/workspace":
			writeTaskContextFixture(t, w, map[string]any{"id": "task", "team": map[string]any{"id": "team", "workspace": map[string]any{"id": "workspace"}}})
		case "/teams/team/components/task":
			writeTaskContextFixture(t, w, map[string]any{"id": "task", "componentType": "task", "title": "Task", "document": map[string]any{"HTMLContent": "<p>body</p>"}})
		case "/teams/team/components/task/relations/issues":
			writeTaskContextFixture(t, w, []map[string]any{
				{"relationCategory": "blockedBy", "issueType": "component", "issue": map[string]any{"id": "rel", "component": map[string]any{"id": "task-b", "title": "Blocking task", "isDone": true}}},
				{"relationCategory": "related", "issueType": "backlog", "issue": map[string]any{"id": "rel2", "backlog": map[string]any{"id": "backlog-1", "title": "Idea"}}},
			})
		case "/teams/team/components/task/relations/meetings":
			http.Error(w, "forbidden", http.StatusForbidden)
		case "/teams/team/components/task/comments":
			page := r.URL.Query().Get("page")
			commentPages = append(commentPages, page)
			var comments []map[string]any
			for i := range 100 {
				n := (len(commentPages)-1)*100 + i
				if n >= 130 {
					break
				}
				comments = append(comments, map[string]any{"id": fmt.Sprintf("c%03d", n), "status": "active", "createdAt": fmt.Sprintf("2026-10-%02dT00:%02d:00Z", 1+n/60, n%60), "body": "comment", "author": map[string]any{"workspaceProfile": map[string]any{"nickname": "유래"}}})
			}
			if page == "1" {
				comments = append(comments, map[string]any{"id": "gone", "status": "deleted", "createdAt": "2026-10-30T00:00:00Z", "body": "x"})
			}
			next := any(nil)
			if page == "1" {
				next = 2
			}
			writeTaskContextFixture(t, w, map[string]any{"comments": comments, "meta": map[string]any{"total": 130, "nextPage": next}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	got, err := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "jwt", Extended: true})
	if err != nil || got.Extended == nil {
		t.Fatalf("context=%+v err=%v", got, err)
	}
	ext := got.Extended
	if len(ext.RelatedWork) != 2 || ext.RelatedWork[0].Title != "Blocking task" || !ext.RelatedWork[0].IsDone || ext.RelatedWork[1].Kind != "backlog" || ext.RelatedStatus != "partial" {
		t.Fatalf("related = %+v status=%s", ext.RelatedWork, ext.RelatedStatus)
	}
	if len(ext.Comments) != 130 || ext.CommentTotal != 130 || ext.CommentStatus != "available" || strings.Join(commentPages, ",") != "1,2" {
		t.Fatalf("comments=%d total=%d status=%s pages=%v", len(ext.Comments), ext.CommentTotal, ext.CommentStatus, commentPages)
	}
	if ext.Comments[0].ID != "c129" || ext.Comments[0].Priority != TaskContextCommentPriorityRecent || ext.Comments[20].Priority != TaskContextCommentPriorityOlder || ext.Comments[0].Author != "유래" {
		t.Fatalf("newest first with priorities: %+v / %+v", ext.Comments[0], ext.Comments[20])
	}
	plain, _ := client.GetAIAgentTaskContextForRequest(context.Background(), AIAgentTaskContextRequest{ComponentID: "task", WorkspaceID: "workspace", BearerToken: "jwt"})
	if plain.Extended != nil || len(commentPages) != 2 {
		t.Fatal("without the flag no extended reads may happen")
	}
}
