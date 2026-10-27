package riidoaiserver

import (
	"encoding/json"
	"os"
	"testing"
)

func TestTaskContextManifestBudgetsMatchRuntime(t *testing.T) {
	data, err := os.ReadFile("../../docs/30-architecture/agent-task-context.riido.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Policy struct {
			SourceBudget map[string]int `json:"source_budget"`
		} `json:"policy"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	expected := map[string]int{"task_document_runes": taskDocumentContextLimit, "parent_document_runes": parentDocumentContextLimit, "conversation_runes": conversationContextLimit, "related_meetings": taskContextMaxMeetings, "meeting_body_bytes": taskContextMeetingBodyBytes, "meeting_concurrency": taskContextMeetingConcurrency, "parents": 3, "recent_messages": 8, "resource_items_per_kind": 8}
	for key, want := range expected {
		if got := manifest.Policy.SourceBudget[key]; got != want {
			t.Errorf("SSOT budget %s = %d, runtime %d", key, got, want)
		}
	}
}
