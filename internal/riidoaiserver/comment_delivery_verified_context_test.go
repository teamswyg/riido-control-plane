package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentDeliveryPreservesVerifiedExtendedContext(t *testing.T) {
	key := strings.Repeat("k", 32)
	evidence := verifiedCommentDelivery{EventID: "event", CommentID: "trigger", TaskID: "task", AgentID: "agent-public-openclaw", WorkspaceID: "workspace-dev-riid", PrincipalID: "user-1", Body: "요약 부탁", Context: AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task", ComponentType: "task", Title: "검토"}, Document: AIAgentTaskContextDocument{Content: "작업 본문"},
		Resources: []AIAgentTaskContextResource{{ID: "link", Kind: "link", Title: "자료", URL: "https://example.test/spec"}}, ResourceStatus: "available",
		Extended: &AIAgentTaskContextExtended{
			RelatedStatus: "available", CommentStatus: "available", CommentTotal: 2,
			Comments:        []AIAgentTaskContextComment{{ID: "parent", Body: "기존 의견", Priority: 2}, {ID: "trigger", ParentID: "parent", Body: "요약 부탁", Priority: 1}},
			RelatedWork:     []AIAgentTaskContextRelatedWork{{ID: "related", Kind: "task", Title: "연관 작업"}},
			RelatedMeetings: []AIAgentTaskContextRelatedMeeting{{ID: "meeting", Title: "연관 미팅", BodyStatus: "available", Document: &AIAgentTaskContextDocument{Content: "회의 본문", ContentFormat: "markdown", Status: "available"}}},
		},
	}}
	evidenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if acceptCommentStatusReaction(w, r) {
			return
		}
		if r.Header.Get(CommentDeliveryKeyHeader) != key || r.Header.Get("Authorization") != "" {
			t.Error("evidence must use existing server key without user bearer")
		}
		_ = json.NewEncoder(w).Encode(evidence)
	}))
	defer evidenceServer.Close()
	agents := NewDevelopmentAIAgentClientStore()
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: agents})
	defer assignments.Close()
	contexts := NewMemoryAssignmentContextStore()
	handler := NewServer(ServerConfig{Assignment: assignments, AIAgentClient: agents, AssignmentContext: contexts, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: evidenceServer.URL, Key: key}}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/internal/agent-comment-deliveries", bytes.NewBufferString(`{"event_id":"event"}`))
	req.Header.Set(CommentDeliveryKeyHeader, key)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	var result AIAgentTaskActionResponse
	if response.Code != http.StatusAccepted || json.Unmarshal(response.Body.Bytes(), &result) != nil {
		t.Fatalf("delivery=%d %s", response.Code, response.Body.String())
	}
	snapshot, found, err := contexts.LoadAssignmentContextSnapshot(t.Context(), result.AssignmentID)
	if err != nil || !found || len(snapshot.Context.Extended.RelatedWork) != 1 || len(snapshot.Context.Extended.RelatedMeetings) != 1 || len(snapshot.Context.Resources) != 1 {
		t.Fatalf("verified references lost: found=%v err=%v snapshot=%+v", found, err, snapshot)
	}
	page, found, err := contexts.LoadAssignmentContextCommentPage(t.Context(), result.AssignmentID, 0)
	if err != nil || !found || len(page) != 2 || page[0].ID != "trigger" || page[0].ParentID != "parent" || page[1].ID != "parent" {
		t.Fatalf("verified comment context lost: found=%v err=%v comments=%+v", found, err, page)
	}
}
