package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func readScopeCommentDeliveryFixture(t *testing.T) verifiedCommentDelivery {
	t.Helper()
	scope, meetings := readScopeWholeFixture()
	meetingPages := []any{}
	for _, ref := range scope["meetings"].([]any) {
		id := ref.(map[string]any)["id"].(string)
		meetingPages = append(meetingPages, map[string]any{"requestCursor": nil, "snapshot": meetings[id]})
	}
	encoded, _ := json.Marshal(map[string]any{"schemaVersion": "riido-read-scope-capture.v1", "scope": scope, "bodyPages": []any{}, "commentPages": []any{}, "meetingPages": meetingPages, "complete": true})
	var capture taskReadScopeEvidence
	if json.Unmarshal(encoded, &capture) != nil {
		t.Fatal("fixture invalid")
	}
	return verifiedCommentDelivery{EventID: "event", CommentID: "comment", TaskID: "task-private", AgentID: "agent-public-openclaw", WorkspaceID: "workspace-dev-riid", PrincipalID: "user-1", Body: "A안과 B안의 장단점 표로 비교 부탁. 비용과 접근성이 기준.", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task-private", ComponentType: "task", Title: "마케팅 방향"}, Document: AIAgentTaskContextDocument{Content: "legacy full body"}, ReadScopeEvidence: &capture}}
}
