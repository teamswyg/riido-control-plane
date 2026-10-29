package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestTaskReadScopeCommentEvidenceMaterializesWithoutRequesterToken(t *testing.T) {
	initial := readScopeCaptureInitial()
	version := initial["works"].([]any)[0].(map[string]any)["body"].(map[string]any)["version"]
	last := readScopeCaptureBody(5, "world", nil, true)
	last["version"] = version
	raw, _ := json.Marshal(map[string]any{"schemaVersion": "riido-read-scope-capture.v1", "scope": initial, "bodyPages": []any{map[string]any{"resourceKind": "component", "page": last}}, "commentPages": []any{}, "meetingPages": []any{}, "complete": true})
	var bundle taskReadScopeEvidence
	if json.Unmarshal(raw, &bundle) != nil {
		t.Fatal("fixture invalid")
	}
	evidence := verifiedCommentDelivery{TaskID: "task-private", PrincipalID: "requester", CommentID: "trigger", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task"}, ReadScopeEvidence: &bundle}}
	got, err := materializeCommentReadScope(t.Context(), evidence)
	if err != nil || got.ReadScope == nil || got.readScopeCapture == nil || got.ReadScopeEvidence != nil {
		t.Fatalf("offline capture=%+v err=%v", got.ReadScope, err)
	}
	body := ""
	for _, page := range got.readScopeCapture.Pages {
		if page.Part == AssignmentContentEditor {
			var value taskReadBodyPage
			_ = json.Unmarshal(page.Data, &value)
			body += value.Content
		}
	}
	if body != "helloworld" {
		t.Fatalf("captured evidence body=%q", body)
	}
	bundle.BodyPages[0].ResourceKind = AssignmentContentBacklog
	if _, err := materializeCommentReadScope(t.Context(), evidence); err == nil {
		t.Fatal("evidence kind spoof accepted")
	}
}
