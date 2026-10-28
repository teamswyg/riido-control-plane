package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type coverageLifecycleFixture struct {
	dynamo     *coverageLifecycleDynamo
	endpoint   string
	devices    *PersistentAIAgentClientStore
	operations *DynamoDBAssignmentOperationStore
	enrollment EnrollDeviceResponse
	cli        *cliInstallationFixture
	binding    AgentRuntimeBinding
	authorizer RequestAuthorizer
}

func newCoverageLifecycleFixture(t *testing.T, provider string) *coverageLifecycleFixture {
	t.Helper()
	dynamo, endpoint := newCoverageLifecycleDynamo(t)
	devices := openCoverageEnrollmentFixture(t, endpoint, true)
	auth := aiAgentClientHTTPAuthorizer(t, []string{"riido:*"}, "owner")
	handler := NewServer(ServerConfig{AIAgentClient: devices, Authorizer: auth}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/v2/desktop/workspaces/workspace-alpha/devices/enroll", bytes.NewBufferString(`{"machine_id":"lifecycle-machine","execution_coverage_protocol":"riido-execution-coverage.v1"}`))
	req.Header.Set(aiAgentTokenHeader, "ai-agent-token")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	var enrolled EnrollDeviceResponse
	if recorder.Code != http.StatusCreated || json.Unmarshal(recorder.Body.Bytes(), &enrolled) != nil || enrolled.ExecutionCoverageGeneration == "" {
		t.Fatalf("actual covered enrollment status=%d", recorder.Code)
	}
	operations, err := NewExecutionCoverageAssignmentStore(devices.snapshotStore, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = operations.Close() })
	binding := AgentRuntimeBinding{AgentID: "agent-owned-codex", DeviceID: enrolled.DeviceID, DaemonID: "coverage-daemon", RuntimeID: "coverage-runtime", RuntimeProvider: provider}
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{binding})
	if err != nil {
		t.Fatal(err)
	}
	cli := newCLIInstallationFixture(t, StoreConfig{OperationStore: operations, AgentRegistry: registry})
	cli.request.ExpectedDeviceID = enrolled.DeviceID
	cli.handler = NewServer(ServerConfig{Assignment: cli.actor, AIAgentClient: devices, Authorizer: auth, CLIInstallationService: cli.service}).Handler()
	return &coverageLifecycleFixture{dynamo, endpoint, devices, operations, enrolled, cli, binding, auth}
}
