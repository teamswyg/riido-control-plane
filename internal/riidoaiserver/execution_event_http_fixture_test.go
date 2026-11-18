package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func reportFixtureEvent(server Server, agentID, deviceID, secret string, event AgentEventRequest) *httptest.ResponseRecorder {
	data, _ := json.Marshal(event)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
	r.Header.Set(deviceIDHeader, deviceID)
	r.Header.Set(deviceSecretHeader, secret)
	w := httptest.NewRecorder()
	server.handleAgentEvent(w, r, agentID)
	return w
}

// Match production coverage admission capability while retaining the ordinary
// (not installation-covered) authenticated device used by this regression.
type reportingOperationStore struct {
	runtimeFakeAssignmentOperationStore
}

func (*reportingOperationStore) loadExecutionCoverage(context.Context, string) (executionCoverage, bool, error) {
	return executionCoverage{}, false, nil
}

func (*reportingOperationStore) authorizeCoveredAssignment(context.Context, executionCoverage, Assignment) error {
	return nil
}
