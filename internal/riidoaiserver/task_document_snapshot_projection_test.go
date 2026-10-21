package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestEditorSnapshotClearsProjectionStatusAndAlternateEvidence(t *testing.T) {
	for _, body := range []string{"<p>Fresh editor body</p>", ""} {
		reader := &assignmentHTTPRequestTaskContextReader{contextSnapshot: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{ID: "task", Title: "Verified task"},
			Document:  AIAgentTaskContextDocument{ID: "document", Content: "Old primary", Source: "provider_json_projection", Status: "empty", AlternateContent: "Old alternate", AlternateFormat: "html", AlternateSource: "component_snapshot"},
		}}
		server := NewServer(ServerConfig{TaskContext: reader})
		result, err := server.assignRequestWithTaskContextPromptForClientResult(context.Background(), "task", AssignRequest{}, AIAgentTaskContextRequest{ComponentID: "task", DocumentSnapshot: &AIAgentTaskDocumentSnapshot{Content: body, ContentFormat: "html"}})
		if err != nil {
			t.Fatal(err)
		}
		prompt := result.Request.Prompt
		if strings.Contains(prompt, "Old primary") || strings.Contains(prompt, "Old alternate") || strings.Contains(prompt, "freshness is unconfirmed") {
			t.Fatalf("stale projection leaked: %s", prompt)
		}
		if body != "" && !strings.Contains(prompt, "Fresh editor body") {
			t.Fatal("fresh body hidden by inherited status")
		}
		if body == "" && !strings.Contains(prompt, "empty when this execution was requested") {
			t.Fatalf("empty editor input treated as unavailable: %s", prompt)
		}
	}
}
