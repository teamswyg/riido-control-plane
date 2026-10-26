package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func extendedAssignmentContextForTest(comments int) AIAgentTaskContext {
	c := assignmentContextTestTaskContext()
	c.Extended = &AIAgentTaskContextExtended{
		RelatedWork:     []AIAgentTaskContextRelatedWork{{ID: "task-b", Kind: "task", Title: "API 계약 정리", Relation: "blockedBy"}},
		RelatedMeetings: []AIAgentTaskContextRelatedMeeting{{ID: "meeting-1", Title: "주간 회의"}},
		RelatedStatus:   "available", CommentStatus: "available", CommentTotal: comments,
	}
	for i := range comments {
		c.Extended.Comments = append(c.Extended.Comments, AIAgentTaskContextComment{ID: fmt.Sprintf("comment-%02d", i), Priority: TaskContextCommentPriorityRecent, Body: fmt.Sprintf("댓글 %d", i), UpdatedAt: "2026-10-25T00:00:00Z"})
	}
	return c
}

func TestAssignmentContextServesCommentPagesAndRelatedTitles(t *testing.T) {
	f := newAssignmentContextFixtureWith(t, NewMemoryAssignmentContextStore(), extendedAssignmentContextForTest(25))
	snapshot, _, _ := f.contexts.LoadAssignmentContextSnapshot(context.Background(), f.leased.ID)
	if snapshot.CommentPages != 2 || snapshot.Context.Extended == nil || len(snapshot.Context.Extended.Comments) != 0 {
		t.Fatalf("comments must be stored as pages outside the snapshot: pages=%d extended=%+v", snapshot.CommentPages, snapshot.Context.Extended)
	}
	reads, _ := f.contexts.ListAssignmentContextReads(context.Background(), f.leased.ID)
	if got := assignmentContextReadSummary(reads); !strings.Contains(got, "prompt:project/summary") || !strings.Contains(got, "prompt:related_task/title") || strings.Contains(got, "comment") {
		t.Fatalf("prompt reads = %s", got)
	}
	first := f.readPage(t, "")
	if len(first.Comments) != taskContextCommentPageSize || first.NextCursor != "1" || first.Total != 25 {
		t.Fatalf("first page = %+v", first)
	}
	second := f.readPage(t, first.NextCursor)
	if len(second.Comments) != 5 || second.NextCursor != "" {
		t.Fatalf("second page = %+v", second)
	}
	reads, _ = f.contexts.ListAssignmentContextReads(context.Background(), f.leased.ID)
	if last := reads[len(reads)-1]; last.SourceKind != "comment" || last.Via != AssignmentContextViaCLI || last.SourceUpdatedAt == "" || last.SourceVersion == "" {
		t.Fatalf("comment read = %+v", last)
	}
	if code, body := f.read(AssignmentContextViewResources, "", f.leased.LeaseToken); code != http.StatusOK || !strings.Contains(body, "API 계약 정리") || !strings.Contains(body, "related_meeting") {
		t.Fatalf("relations status=%d body=%s", code, body)
	}
}

func (f *assignmentContextFixture) readPage(t *testing.T, cursor string) AssignmentContextCommentPage {
	t.Helper()
	body, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: AssignmentContextViewComments, Target: "RIID-5900", Cursor: cursor})
	code, resp := f.do("/v1/agents/agent-owned-codex/assignment-context", string(body), f.device)
	var out struct{ Data AssignmentContextCommentPage }
	if code != http.StatusOK || json.Unmarshal([]byte(resp), &out) != nil {
		t.Fatalf("comments status=%d body=%s", code, resp)
	}
	return out.Data
}
