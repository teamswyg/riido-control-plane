package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExplicitNativeViewerMetadataDoesNotCapture(t *testing.T) {
	calls := 0
	missing := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer viewer" {
			t.Error("wrong viewer")
		}
		var in struct {
			Query     string
			Variables map[string]any
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if strings.Contains(in.Query, "getComponentWorkspace") {
			_, _ = w.Write([]byte(`{"data":{"getComponentWorkspace":{"id":"target","componentType":"task","team":{"id":"team","workspace":{"id":"workspace"}}}}}`))
			return
		}
		if !strings.Contains(in.Query, "metadata:getComponent(") || strings.Contains(in.Query, "getComponentReadScope") || strings.Contains(in.Query, "HTMLContent") || strings.Contains(in.Query, "description") || in.Variables["workspace"] != "workspace" {
			t.Error("native receipt fetched capture or omitted workspace")
			http.Error(w, "invalid", http.StatusBadRequest)
			return
		}
		if missing {
			_, _ = w.Write([]byte(`{"errors":[{"extensions":{"code":"NOT_FOUND"}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"metadata":{"id":"target","title":"Current title","componentType":"task"}}}`))
	}))
	defer upstream.Close()
	c, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2, AssignmentReadScope: true})
	if err != nil {
		t.Fatal(err)
	}
	req := AIAgentTaskContextRequest{ComponentID: "target", WorkspaceID: "workspace", BearerToken: "viewer", metadataOnly: true}
	out, err := c.GetAIAgentTaskContextForRequest(t.Context(), req)
	if err != nil || out.Component.Title != "Current title" || calls != 2 {
		t.Fatalf("metadata failed: %v calls=%d", err, calls)
	}
	missing = true
	s := Server{taskContext: c}
	if _, visible, err := s.currentExplicitReadMetadata(t.Context(), "viewer", "workspace", runReadSourceKey{"task", "target", "metadata"}); err != nil || visible {
		t.Fatalf("deleted source must hide: %v %v", visible, err)
	}
}
