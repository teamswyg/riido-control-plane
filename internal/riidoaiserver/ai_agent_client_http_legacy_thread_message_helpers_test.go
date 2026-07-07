package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type legacyThreadMessageStore struct {
	*DevelopmentAIAgentClientStore
	thread  AIAgentTaskThreadRecord
	findErr error
}

func (s legacyThreadMessageStore) FindAIAgentTaskThreadByID(
	context.Context,
	string,
	string,
) (AIAgentTaskThreadRecord, error) {
	if s.findErr != nil {
		return AIAgentTaskThreadRecord{}, s.findErr
	}
	return s.thread, nil
}

func postLegacyThreadMessage(
	t *testing.T,
	server http.Handler,
	path string,
) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(CreateAIAgentTaskThreadMessageRequest{Body: "continue"})
	if err != nil {
		t.Fatalf("Marshal message request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer user-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	return resp
}
