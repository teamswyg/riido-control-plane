package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTaskReadScopeClickSelectionPreservesSavedBodyContinuation(t *testing.T) {
	const saved = "<p>Saved body</p>"
	const editor = "<p>Unsaved selection</p>"
	initial := readScopeCaptureInitial()
	work := initial["works"].([]any)[0].(map[string]any)
	first := readScopeWholeBody("task-private", AssignmentContentComponent, "document", saved)
	first["content"], first["nextOffset"], first["complete"], first["status"] = saved[:8], 8, false, "partial"
	work["body"] = first
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string
			Variables map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		calls.Add(1)
		root, response := "getComponentReadScope", any(initial)
		if strings.Contains(req.Query, "getComponentReadScopeBody(") {
			if req.Variables["version"] != first["version"] || req.Variables["offset"] != float64(8) {
				t.Error("saved body continuation lost its version or offset")
			}
			page := readScopeWholeBody("task-private", AssignmentContentComponent, "document", saved)
			page["content"], page["offset"] = saved[8:], 8
			root, response = "getComponentReadScopeBody", page
		} else {
			snapshot, ok := req.Variables["editorSnapshot"].(map[string]any)
			if !ok || snapshot["content"] != editor {
				t.Error("click snapshot was not passed for reference selection")
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: response}})
	}))
	t.Cleanup(server.Close)
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
	got, err := client.getTaskReadScope(t.Context(), AIAgentTaskContextRequest{ComponentID: "task-private", BearerToken: "synthetic", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: editor, ContentFormat: "html"}}, "team")
	if err != nil {
		t.Fatal("producer-authorized saved document was rejected", err)
	}
	var restored strings.Builder
	for _, page := range got.readScopeCapture.Pages {
		if page.Part == AssignmentContentEditor {
			var body taskReadBodyPage
			if err := json.Unmarshal(page.Data, &body); err != nil {
				t.Fatal(err)
			}
			if body.Source != "snapshot_html" || body.Version != first["version"] {
				t.Fatal("saved document provenance changed")
			}
			restored.WriteString(body.Content)
		}
	}
	if calls.Load() != 2 || restored.String() != saved {
		t.Fatalf("saved document continuation changed: calls=%d body=%q", calls.Load(), restored.String())
	}
}

func TestTaskReadScopeClientSnapshotStillRequiresExactInput(t *testing.T) {
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "task-private"}
	body := &taskReadBodyPage{Source: "client_snapshot", Part: "document", Format: "html", Content: "<p>Forged</p>", Version: "wrong-version"}
	for _, editor := range []*AIAgentTaskDocumentSnapshot{nil, {Content: "<p>Actual</p>", ContentFormat: "html"}} {
		if validateReadScopeEditor(resource, body, editor) == nil {
			t.Fatal("unbound client snapshot accepted")
		}
	}
}
