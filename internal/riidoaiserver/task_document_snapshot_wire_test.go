package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestTaskDocumentSnapshotWireRequiresExplicitContentAndFormat(t *testing.T) {
	for _, input := range []string{`{"document_snapshot":{}}`, `{"document_snapshot":{"content_format":"html"}}`, `{"document_snapshot":{"content":null,"content_format":"html"}}`, `{"document_snapshot":{"content":"","content_format":"markdown"}}`, `{"document_snapshot":{"content":"","content_format":"html","workspace_id":"foreign"}}`} {
		var request AssignAIAgentTaskRequest
		if err := json.Unmarshal([]byte(input), &request); err == nil {
			t.Fatalf("accepted invalid wire snapshot %s", input)
		}
	}
	var request AssignAIAgentTaskRequest
	if err := json.Unmarshal([]byte(`{"document_snapshot":{"content":"","content_format":"html"}}`), &request); err != nil || request.DocumentSnapshot == nil || request.DocumentSnapshot.Content != "" {
		t.Fatal("explicit empty body lost", err)
	}
}
