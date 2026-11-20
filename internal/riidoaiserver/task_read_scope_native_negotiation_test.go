package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskReadScopeNativeInternalNegotiationRejectsDowngrade(t *testing.T) {
	for _, schema := range []string{taskReadScopeNativeSchema, taskReadScopeSchema, ""} {
		t.Run(schema, func(t *testing.T) {
			bundle := nativeTestCapture(t, "backlog")
			bundle.Scope.SchemaVersion = schema
			evidence := verifiedCommentDelivery{TaskID: "task-private", WorkspaceID: "workspace", PrincipalID: "requester", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task", Kind: "backlog"}, ReadScopeEvidence: bundle}}
			key := strings.Repeat("k", 32)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get(commentReadScopeVersionHeader) != taskReadScopeNativeSchema || r.Header.Get(CommentDeliveryKeyHeader) != key {
					t.Error("native evidence negotiation/auth lost")
				}
				_ = json.NewEncoder(w).Encode(evidence)
			}))
			defer upstream.Close()
			private, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
			server := Server{taskContext: private, config: ServerConfig{AssignmentReadScope: true, CommentDelivery: CommentDeliveryConfig{Key: key}}}
			got, err := server.fetchInvocationEvidence(t.Context(), upstream.URL)
			if schema != taskReadScopeNativeSchema {
				if err == nil {
					t.Fatal("native negotiation downgraded")
				}
				return
			}
			if err != nil || got.Context.Component.Kind != "backlog" || got.Context.ReadScope.WorkspaceID != "workspace" {
				t.Fatalf("native evidence: %+v %v", got, err)
			}
		})
	}
}
