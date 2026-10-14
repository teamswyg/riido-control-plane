package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPBootstrapCoalescesWorkspaceWideReconcile(t *testing.T) {
	server, aiAgentStore := newRecordingAIAgentAssignmentTestServer(t)

	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("bootstrap status=%d body=%s", resp.Code, resp.Body.String())
		}
	}

	assertStringSequence(t, "reconcile task ids", aiAgentStore.reconcileTaskIDs(), "")
}
