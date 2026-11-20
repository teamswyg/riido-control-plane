package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskReadScopeNativeTypedMetadataAndCommentsOfflineReplay(t *testing.T) {
	prefixes, payloads, evidence := nativeTypedFixture()
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables struct {
				Part           string
				Offset         int
				WorkspaceID    string
				LocatorVersion int
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if req.Variables.WorkspaceID != "workspace" || req.Variables.LocatorVersion != 2 {
			t.Error("native metadata locator lost")
		}
		calls++
		root, response := "getComponentReadScope", any(evidence.Scope)
		if req.Variables.Part != "" {
			first := prefixes[req.Variables.Part]
			page := typedJSONFixture(first.ResourceID, first.Part, first.Version, payloads[first.Part], req.Variables.Offset, first.Selector)
			evidence.BodyPages = append(evidence.BodyPages, taskReadCapturedBody{ResourceKind: AssignmentContentComponent, Page: page})
			root, response = "getComponentReadScopeBody", page
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: response}})
	}))
	defer upstream.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
	req := AIAgentTaskContextRequest{ComponentID: "task-private", WorkspaceID: "workspace", BearerToken: "human"}
	live, err := client.getTaskReadScope(t.Context(), req, "team")
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range live.readScopeCapture.Manifest {
		if !part.Complete {
			t.Fatalf("incomplete typed native capture: %+v", part)
		}
	}
	index, err := newTaskReadEvidenceIndex(evidence, req.ComponentID, "team")
	if err != nil {
		t.Fatal(err)
	}
	client.readScopeEvidence = index
	before := calls
	offline, err := client.getTaskReadScope(t.Context(), req, "team")
	if err != nil || calls != before || len(offline.readScopeCapture.Pages) != len(live.readScopeCapture.Pages) {
		t.Fatalf("native replay differed: %v", err)
	}
	for i, page := range live.readScopeCapture.Pages {
		if string(page.Data) != string(offline.readScopeCapture.Pages[i].Data) {
			t.Fatalf("native replay changed page %d", i)
		}
	}
}
