package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskReadScopeRejectsInitialPartOrRequiredListMismatch(t *testing.T) {
	for _, field := range []string{"body-part", "description-part", "alternate-part", "priorities", "checklists", "attachments"} {
		t.Run(field, func(t *testing.T) {
			input := readScopeCaptureInitial()
			work := input["works"].([]any)[0].(map[string]any)
			switch field {
			case "body-part":
				work["body"].(map[string]any)["part"] = "description"
			case "description-part":
				work["descriptionBody"] = readScopeWholeBody("task-private", AssignmentContentComponent, "document", "wrong-slot")
			case "alternate-part":
				work["alternateBody"] = readScopeWholeBody("task-private", AssignmentContentComponent, "description", "wrong-slot")
			default:
				delete(work, field)
			}
			raw, _ := json.Marshal(input)
			var scope taskReadScope
			_ = json.Unmarshal(raw, &scope)
			if validateTaskReadScope(scope, "task-private", "team") == nil {
				t.Fatal("malformed first page/list accepted")
			}
		})
	}
}

func TestTaskReadScopeMeetingRetainsMetadataPartialAndRejectsInvalidStatus(t *testing.T) {
	for _, status := range []string{"partial", "invalid"} {
		t.Run(status, func(t *testing.T) {
			_, meetings := readScopeWholeFixture()
			meeting := meetings["meeting-0"].(map[string]any)
			meeting["partsStatus"] = []any{map[string]any{"name": "labels", "status": status, "complete": false}}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"getMeetingReadScope": meeting}})
			}))
			t.Cleanup(server.Close)
			client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
			b, _ := newTaskReadScopeBuilder()
			resource, err := client.captureReadScopeMeeting(t.Context(), "synthetic", b, taskReadScopeMeetingRef{ID: "meeting-0", TeamID: "team"})
			if status == "invalid" {
				if err == nil {
					t.Fatal("invalid source status accepted")
				}
				return
			}
			if err != nil || len(resource.Parts) == 0 || resource.Parts[0].Complete {
				t.Fatalf("metadata failure hidden: %+v %v", resource, err)
			}
		})
	}
}
