package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTaskReadScopeDeadlineKeepsCapturedPrefixWithoutInventedCursor(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			_, _ = w.Write([]byte(`{"id":"task-private","team":{"id":"team","workspace":{"id":"workspace-dev-riid"}}}`))
			return
		}
		calls.Add(1)
		var req struct{ Query string }
		_ = json.NewDecoder(r.Body).Decode(&req)
		if strings.Contains(req.Query, "getComponentReadScopeBody") {
			<-r.Context().Done()
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": readScopeCaptureInitial()}})
	}))
	t.Cleanup(server.Close)
	c, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL, AssignmentReadScope: true, Timeout: 250 * time.Millisecond})
	got, err := c.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace-dev-riid", BearerToken: "synthetic"})
	if err != nil || calls.Load() != 2 || got.ReadScope == nil || got.ReadScope.Status != "partial" {
		t.Fatalf("deadline capture calls=%d scope=%+v err=%v", calls.Load(), got.ReadScope, err)
	}
	for _, part := range got.readScopeCapture.Manifest {
		if part.Part == AssignmentContentEditor && (part.Complete || len(part.Cursors) != 1) {
			t.Fatalf("deadline fabricated completion/cursor: %+v", part)
		}
	}
}

func TestTaskReadScopeCaptureBudgetPublishesOnlyStoredPages(t *testing.T) {
	b, _ := newTaskReadScopeBuilder()
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work"}
	raw, _ := json.Marshal(map[string]string{"body": strings.Repeat("large", 250000)})
	pages := readScopeJSONPages(raw, 0, 0, []AssignmentContextSource{{Kind: "task", ID: "work", Part: "metadata"}})
	part := b.addPart(resource, AssignmentContentMetadata, "version", pages, true)
	if part.Complete || len(part.Cursors) == 0 || len(part.Cursors) >= len(pages) || part.Status != "partial" {
		t.Fatalf("budget not explicit: %+v", part)
	}
	for _, page := range b.capture.Pages {
		if page.Complete {
			t.Fatal("uncaptured end marked complete")
		}
	}
	snapshot := AssignmentContextSnapshot{AssignmentID: "bounded", TaskID: "work", Content: b.capture.Manifest}
	store := NewMemoryAssignmentContextStore()
	if err := store.SaveAssignmentContentBundle(t.Context(), snapshot, b.capture.Pages); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := store.LoadAssignmentContentPage(t.Context(), snapshot.AssignmentID, len(part.Cursors)); found {
		t.Fatal("uncaptured page addressable")
	}
}
