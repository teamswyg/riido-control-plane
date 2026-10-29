package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentDeliveryReadScopeNegotiationAndBoundedFullCapture(t *testing.T) {
	for _, scenario := range []string{"scoped", "missing", "legacy"} {
		t.Run(scenario, func(t *testing.T) {
			key := strings.Repeat("k", 32)
			evidence := readScopeCommentDeliveryFixture(t)
			if scenario == "missing" {
				evidence.Context.ReadScopeEvidence = nil
			}
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get(CommentDeliveryKeyHeader) != key {
					t.Error("service evidence auth missing")
				}
				want := taskReadScopeSchema
				if scenario == "legacy" {
					want = ""
				}
				// RIID-5907 reaction writes share this upstream; only evidence reads negotiate scope.
				if r.Method == http.MethodGet && r.Header.Get(commentReadScopeVersionHeader) != want {
					t.Errorf("capability=%q want%q", r.Header.Get(commentReadScopeVersionHeader), want)
				}
				_ = json.NewEncoder(w).Encode(evidence)
			}))
			t.Cleanup(upstream.Close)
			client := NewDevelopmentAIAgentClientStore()
			assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: client})
			t.Cleanup(func() { assignments.Close() })
			contexts := NewMemoryAssignmentContextStore()
			handler := NewServer(ServerConfig{AIAgentClient: client, Assignment: assignments, AssignmentContext: contexts, AssignmentReadScope: scenario != "legacy", CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: upstream.URL, Key: key}}).Handler()
			request := httptest.NewRequest(http.MethodPost, "/internal/agent-comment-deliveries", bytes.NewBufferString(`{"event_id":"event"}`))
			request.Header.Set(CommentDeliveryKeyHeader, key)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if scenario == "missing" {
				if response.Code != http.StatusBadGateway {
					t.Fatalf("missing capture accepted: %d", response.Code)
				}
				return
			}
			var action AIAgentTaskActionResponse
			if response.Code != http.StatusAccepted || json.Unmarshal(response.Body.Bytes(), &action) != nil || action.AssignmentID == "" {
				t.Fatalf("delivery=%d %s", response.Code, response.Body.String())
			}
			snapshot, found, err := contexts.LoadAssignmentContextSnapshot(t.Context(), action.AssignmentID)
			if err != nil || !found {
				t.Fatalf("snapshot=%v %v", found, err)
			}
			if scenario == "legacy" {
				if snapshot.Context.ReadScope != nil || snapshot.Context.ReadScopeEvidence != nil || snapshot.Context.Document.Content != "legacy full body" {
					t.Fatal("legacy capture changed")
				}
				return
			}
			if snapshot.Truncated || len(snapshot.Content) != 120 || len(snapshot.Context.ReadScope.Resources) != 30 {
				t.Fatalf("comment scope incomplete: parts=%d truncated=%v", len(snapshot.Content), snapshot.Truncated)
			}
			for i := range snapshot.Content {
				if _, found, err := contexts.LoadAssignmentContentPage(t.Context(), action.AssignmentID, i); err != nil || !found {
					t.Fatalf("comment page%d=%v %v", i, found, err)
				}
			}
		})
	}
}
