package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"unicode/utf8"
)

func TestTaskReadScopeLargeClickSnapshotUsesOnlyValidatedLocalContinuation(t *testing.T) {
	content := "<p>" + strings.Repeat("한글😀", 12000) + "</p>"
	end := 8192
	for !utf8.RuneStart(content[end]) {
		end--
	}
	initial := readScopeCaptureInitial()
	work := initial["works"].([]any)[0].(map[string]any)
	first := readScopeWholeBody("task-private", AssignmentContentComponent, "document", content)
	first["source"] = "client_snapshot"
	first["version"] = readScopeBodyVersion(AssignmentContentComponent, "task-private", "document", "client_snapshot", "html", content)
	first["content"], first["nextOffset"], first["complete"], first["status"] = content[:end], end, false, "partial"
	work["body"] = first
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/workspace") {
			_, _ = w.Write([]byte(`{"id":"task-private","team":{"id":"team","workspace":{"id":"workspace-dev-riid"}}}`))
			return
		}
		calls.Add(1)
		var req struct {
			Query     string
			Variables map[string]any
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if calls.Load() != 1 || strings.Contains(req.Query, "getComponentReadScopeBody") {
			t.Error("click snapshot attempted a live continuation")
		}
		editor, ok := req.Variables["editorSnapshot"].(map[string]any)
		if !ok || editor["content"] != content || editor["contentFormat"] != "html" {
			t.Error("click selection did not receive original typed snapshot")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getComponentReadScope": initial}})
	}))
	t.Cleanup(server.Close)
	c, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL, AssignmentReadScope: true})
	got, err := c.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace-dev-riid", BearerToken: "synthetic", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: content, ContentFormat: "html"}})
	if err != nil || got.readScopeCapture == nil {
		t.Fatal("click capture failed", err)
	}
	var joined strings.Builder
	count := 0
	for _, page := range got.readScopeCapture.Pages {
		if page.Part == AssignmentContentEditor {
			var body taskReadBodyPage
			_ = json.Unmarshal(page.Data, &body)
			if !utf8.ValidString(body.Content) {
				t.Fatal("split UTF8")
			}
			joined.WriteString(body.Content)
			count++
		}
	}
	if calls.Load() != 1 || count < 10 || joined.String() != content {
		t.Fatalf("click capture incomplete: calls=%d pages=%d bytes=%d", calls.Load(), count, joined.Len())
	}
}
