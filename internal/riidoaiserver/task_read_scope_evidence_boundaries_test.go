package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestTaskReadScopeEvidenceIncompleteCannotBecomeComplete(t *testing.T) {
	initial, _ := readScopeWholeFixture()
	initial["meetings"] = []any{}
	raw, _ := json.Marshal(map[string]any{"schemaVersion": "riido-read-scope-capture.v1", "scope": initial, "bodyPages": []any{}, "commentPages": []any{}, "meetingPages": []any{}, "complete": false})
	var bundle taskReadScopeEvidence
	_ = json.Unmarshal(raw, &bundle)
	evidence := verifiedCommentDelivery{TaskID: "task-private", PrincipalID: "requester", CommentID: "trigger", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task"}, ReadScopeEvidence: &bundle}}
	got, err := materializeCommentReadScope(t.Context(), evidence)
	if err != nil || got.ReadScope == nil || got.ReadScope.Status != "partial" {
		t.Fatalf("producer incomplete was promoted: scope-present=%v err=%v", got.ReadScope != nil, err)
	}
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	f.contexts.mu.Lock()
	delete(f.contexts.snapshots, f.leased.ID)
	f.contexts.mu.Unlock()
	f.service.config.AssignmentReadScope = true
	f.service.captureAssignmentContext(t.Context(), f.leased, &got, assignmentContextTrigger{Kind: AssignmentContextTriggerComment})
	snapshot, found, err := f.contexts.LoadAssignmentContextSnapshot(t.Context(), f.leased.ID)
	if err != nil || !found || !snapshot.Truncated {
		t.Fatalf("incomplete snapshot advertised complete: found=%v truncated=%v err=%v", found, snapshot.Truncated, err)
	}
}
