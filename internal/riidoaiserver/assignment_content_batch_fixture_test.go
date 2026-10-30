package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Six 3,000-byte body pages followed by a second part of the same source.
func newAssignmentContentBatchFixture(t *testing.T) *assignmentContextFixture {
	t.Helper()
	m := NewMemoryAssignmentContextStore()
	f := newAssignmentContextFixture(t, m)
	snapshot, _, err := m.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	body := AssignmentContentManifest{ResourceKind: AssignmentContentComponent, SourceID: "task-private", Part: AssignmentContentDescription, Version: "v1", Status: "available", Complete: true}
	comments := AssignmentContentManifest{ResourceKind: AssignmentContentComponent, SourceID: "task-private", Part: AssignmentContentComments, Version: "c1", Status: "available", Complete: true, Cursors: []string{"comments-0"}}
	var pages []AssignmentContentPage
	for i := range 6 {
		cursor := fmt.Sprintf("body-%d", i)
		body.Cursors = append(body.Cursors, cursor)
		data, _ := json.Marshal(map[string]string{"content": strings.Repeat(string(rune('a'+i)), 3000)})
		pages = append(pages, AssignmentContentPage{ResourceKind: AssignmentContentComponent, Cursor: cursor, SourceID: "task-private", Part: AssignmentContentDescription, Version: "v1", Offset: i * 3000, Status: "partial", Complete: i == 5, Data: data, Sources: []AssignmentContextSource{{Kind: "task", ID: "task-private", Part: "body", Version: "v1"}}})
	}
	pages[5].Status = "available"
	pages = append(pages, AssignmentContentPage{ResourceKind: AssignmentContentComponent, Cursor: "comments-0", SourceID: "task-private", Part: AssignmentContentComments, Version: "c1", Status: "available", Complete: true, Data: json.RawMessage(`{"items":[]}`), Sources: []AssignmentContextSource{{Kind: "task", ID: "task-private", Part: "comments", Version: "c1"}}})
	snapshot.Content = []AssignmentContentManifest{body, comments}
	m.mu.Lock()
	delete(m.snapshots, f.leased.ID)
	m.mu.Unlock()
	if err := m.SaveAssignmentContentBundle(t.Context(), snapshot, pages); err != nil {
		t.Fatal(err)
	}
	return f
}

type assignmentContentBatchResponse struct {
	Data struct {
		SourceID   string                  `json:"source_id"`
		Part       string                  `json:"part"`
		Version    string                  `json:"version"`
		Pages      []AssignmentContentPage `json:"pages"`
		Complete   bool                    `json:"complete"`
		NextCursor string                  `json:"next_cursor"`
	} `json:"data"`
	Sources []AssignmentContextSource `json:"sources"`
}

func readAssignmentContentBatch(t *testing.T, f *assignmentContextFixture, cursor string, maxBytes int) (int, string) {
	t.Helper()
	request, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: f.leased.LeaseToken, View: "content", Target: "task-private", Cursor: cursor, MaxBytes: maxBytes})
	return f.do("/v1/agents/agent-owned-codex/assignment-context", string(request), f.device)
}
