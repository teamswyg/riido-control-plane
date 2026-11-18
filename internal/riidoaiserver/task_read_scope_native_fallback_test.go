package riidoaiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskReadScopeNativeTransportFailureCannotDispatchFallback(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	upstream.Close()
	client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: upstream.URL, LocatorVersion: 2})
	server := NewServer(ServerConfig{TaskContext: client})
	result, err := server.assignRequestWithTaskContextPromptForClientResult(t.Context(), "task-private", AssignRequest{Prompt: "Inspect only"}, AIAgentTaskContextRequest{WorkspaceID: "workspace", BearerToken: "human"})
	var unavailable latestTaskDocumentError
	if !errors.As(err, &unavailable) || result.Request.Prompt != "" {
		t.Fatalf("transport failure dispatched fallback: %+v, %v", result, err)
	}
}

func TestTaskReadScopeNativeCompositionFailureCannotDispatchFallback(t *testing.T) {
	reader := &assignmentHTTPRequestTaskContextReader{contextSnapshot: AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task-private"},
		ReadScope: &AssignmentReadScope{SchemaVersion: taskReadScopeSchema, SourceSchemaVersion: taskReadScopeNativeSchema, WorkspaceID: "workspace"},
	}}
	server := NewServer(ServerConfig{TaskContext: reader})
	result, err := server.assignRequestWithTaskContextPromptForClientResult(t.Context(), "task-private", AssignRequest{Prompt: "Inspect only"}, AIAgentTaskContextRequest{WorkspaceID: "workspace", BearerToken: "human"})
	if err == nil || result.TaskContext != nil || result.Request.Prompt != "" {
		t.Fatalf("native composition failure dispatched fallback: %+v, %v", result, err)
	}
}
