package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskReadScopeNativeCommentContinuationWorkspace(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string
			Variables map[string]any
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Variables["workspaceId"] != "workspace" || req.Variables["locatorVersion"] != float64(2) || req.Variables["teamId"] != "" || !strings.Contains(req.Query, "workspaceId:$workspaceId") {
			t.Error("comment continuation lost native locator")
		}
		page := readScopeCommentFixture("", 2, []any{}, nil, true)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getReadScopeComments": page}})
	}))
	defer upstream.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
	_, err := client.fetchReadScopeComments(t.Context(), "human", AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work", WorkspaceID: "workspace"}, "", "root-comments", 2, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskReadScopeNativeEditorContinuationHashesWorkspace(t *testing.T) {
	content := strings.Repeat("한글😀", 6000)
	resource := AssignmentScopedResource{WorkspaceID: "workspace", Kind: AssignmentContentIntake, ID: "task", TeamID: "team"}
	version := nativeTestDigest([]string{"riido-read-body.v2", "workspace", "intake", "task", "team", "document", "client_snapshot", "html", content})
	offset, next, complete := 0, len("한글😀"), false
	first := &taskReadBodyPage{ResourceID: "task", Part: "document", Status: "partial", Source: "client_snapshot", Format: "html", Version: version, Offset: &offset, NextOffset: &next, Complete: &complete, Content: content[:next]}
	editor := &AIAgentTaskDocumentSnapshot{Content: content, ContentFormat: "html"}
	if err := validateReadScopeEditor(resource, first, editor); err != nil {
		t.Fatal(err)
	}
	b, _ := newTaskReadScopeBuilder()
	client := &AIAgentPrivateTaskContextClient{}
	part, err := client.captureReadScopeBody(t.Context(), "", b, resource, first, editor)
	if err != nil || !part.Complete || len(part.Cursors) < 2 {
		t.Fatalf("native editor incomplete: %+v %v", part, err)
	}
	resource.WorkspaceID = "other"
	if err := validateReadScopeEditor(resource, first, editor); err == nil {
		t.Fatal("cross-workspace editor hash accepted")
	}
}
