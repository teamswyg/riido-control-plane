package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExplicitReadReceiptRequiresCurrentViewerAccess(t *testing.T) {
	reader := &runReadsTestContext{value: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "target-one", ComponentType: "task", Title: "Current title"}}}
	s := Server{taskContext: reader}
	key := runReadSourceKey{"task", "target-one", "metadata"}
	meta, ok, err := s.currentExplicitReadMetadata(t.Context(), "viewer-token", "workspace-one", key)
	if err != nil || !ok || meta.title != "Current title" || len(reader.requests) != 1 || reader.requests[0].BearerToken != "viewer-token" || reader.requests[0].ComponentID != "target-one" {
		t.Fatal("wrong viewer read", err)
	}
	reader.err = ErrAuthorizationForbidden
	if _, ok, err = s.currentExplicitReadMetadata(t.Context(), "viewer-token", "workspace-one", key); err != nil || ok {
		t.Fatal("hidden receipt made visible")
	}
	visible := map[runReadSourceKey]runReadMetadata{key: {title: "Old visible title"}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer viewer-token")
	if !s.addVisibleExplicitReadSources(httptest.NewRecorder(), req, AuthorizationResult{WorkspaceID: "workspace-one"}, []AssignmentContextRead{{Explicit: true, SourceKind: "task", SourceID: "target-one", SourcePart: "metadata"}}, visible) || len(visible) != 0 {
		t.Fatal("hidden count or metadata exposed")
	}
}

func TestExplicitViewerMetadataSelectsNoBody(t *testing.T) {
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer viewer" {
			t.Error("wrong viewer")
		}
		if r.URL.Path == "/public/components/target/workspace" {
			_, _ = w.Write([]byte(`{"id":"target","team":{"id":"team","workspace":{"id":"workspace"}}}`))
			return
		}
		if r.URL.Path != "/graphql" {
			http.Error(w, strings.Repeat("body", 1<<20), http.StatusOK)
			return
		}
		var in struct {
			Query     string
			Variables map[string]string
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if strings.Contains(in.Query, "description") || strings.Contains(in.Query, "HTMLContent") || in.Variables["id"] != "target" || in.Variables["team"] != "team" {
			t.Error("not bounded metadata")
		}
		_, _ = w.Write([]byte(`{"data":{"metadata":{"id":"target","title":"Current title","componentType":"task"}}}`))
	}))
	defer upstream.Close()
	client, err := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, AssignmentReadScope: true})
	if err != nil {
		t.Fatal(err)
	}
	out, err := client.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "target", WorkspaceID: "workspace", BearerToken: "viewer", metadataOnly: true})
	if err != nil || out.Component.Title != "Current title" || calls != 2 {
		t.Fatalf("bounded metadata failed: %v calls=%d", err, calls)
	}
}
