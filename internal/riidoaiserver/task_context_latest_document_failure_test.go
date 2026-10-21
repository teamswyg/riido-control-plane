package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestProviderDocumentHTTPFailuresAreExplicit(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/documents/providers/team-a/component-a" {
				http.Error(w, "unavailable", status)
				return
			}
			writeProviderDocumentContextFixture(t, w, r)
		}))
		client, _ := NewAIAgentPrivateTaskContextClient(AIAgentPrivateTaskContextClientConfig{BaseURL: server.URL})
		_, err := client.GetAIAgentTaskContextForRequest(t.Context(), AIAgentTaskContextRequest{ComponentID: "component-a", WorkspaceID: "workspace-a", BearerToken: "user-jwt"})
		server.Close()
		var latest latestTaskDocumentError
		if !errors.As(err, &latest) {
			t.Fatalf("HTTP %d silently fell back: %v", status, err)
		}
		response := httptest.NewRecorder()
		writeAIAgentClientError(response, err)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("HTTP failure status=%d", response.Code)
		}
	}
}

func TestLatestDocumentErrorCannotBecomeBodylessAssignment(t *testing.T) {
	s := NewServer(ServerConfig{TaskContext: &assignmentHTTPRequestTaskContextReader{err: latestTaskDocumentError{context.DeadlineExceeded}}})
	_, err := s.assignRequestWithTaskContextPromptForClientResult(t.Context(), "task", AssignRequest{}, AIAgentTaskContextRequest{ComponentID: "task"})
	if err == nil {
		t.Fatal("latest-document failure executed without task body")
	}
}
