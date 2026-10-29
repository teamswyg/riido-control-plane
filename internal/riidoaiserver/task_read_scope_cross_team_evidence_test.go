package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestTaskReadScopeEvidencePreservesSelectedLeafTeam(t *testing.T) {
	scope, meeting, last := readScopeCrossTeamFixture()
	raw, _ := json.Marshal(map[string]any{"schemaVersion": "riido-read-scope-capture.v1", "scope": scope, "bodyPages": []any{map[string]any{"resourceKind": "component", "page": last}}, "commentPages": []any{}, "meetingPages": []any{map[string]any{"requestCursor": nil, "snapshot": meeting}}, "complete": true})
	var bundle taskReadScopeEvidence
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatal(err)
	}
	evidence := verifiedCommentDelivery{TaskID: "task-private", PrincipalID: "requester", CommentID: "trigger", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task"}, ReadScopeEvidence: &bundle}}
	got, err := materializeCommentReadScope(t.Context(), evidence)
	if err != nil || got.ReadScope == nil || got.ReadScope.Status != "available" {
		t.Fatalf("offline leaf capture: %v", err)
	}
	index, err := newTaskReadEvidenceIndex(&bundle, "task-private", "team")
	if err != nil {
		t.Fatal(err)
	}
	variables := map[string]any{"componentId": "work-1", "teamId": "team", "part": "document", "version": last["version"], "offset": 5}
	var body taskReadBodyPage
	if index.read(t.Context(), "getComponentReadScopeBody", variables, &body) == nil {
		t.Fatal("leaf page read through wrong team")
	}
	variables["teamId"] = "leaf-team"
	if err := index.read(t.Context(), "getComponentReadScopeBody", variables, &body); err != nil {
		t.Fatal(err)
	}
	bundle.MeetingPages[0].Snapshot.TeamID = "team"
	if _, err := newTaskReadEvidenceIndex(&bundle, "task-private", "team"); err == nil {
		t.Fatal("meeting changed selected team")
	}
}
