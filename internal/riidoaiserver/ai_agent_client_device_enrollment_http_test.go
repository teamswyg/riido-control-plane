package riidoaiserver

import (
	"net/http"
	"testing"
)

type deviceEnrollmentHTTPFixture struct {
	t               *testing.T
	server          http.Handler
	aiAgentStore    *DevelopmentAIAgentClientStore
	enrollment      EnrollDeviceResponse
	created         AgentClientRecordResponse
	codexRuntimeID  string
	cursorRuntimeID string
}

func TestHTTPDesktopDeviceEnrollmentAndDaemonCredentialAuthorization(t *testing.T) {
	f := newDeviceEnrollmentHTTPFixture(t)
	f.enrollAndRotateDeviceCredential()
	f.verifyEnrolledDevicePrincipal()
	f.verifyDeviceCredentialPollAuthorization()
	f.syncCodexRuntimeAndCreateAgent()
	f.syncCursorRuntimeAndVerifyMergedReadModels()
	f.verifyDaemonBindingsAfterRuntimeSnapshot()
	f.projectStaleRuntimeAndDaemonOffline()
	f.moveRuntimeToReplacementDevice()
}

func newDeviceEnrollmentHTTPFixture(t *testing.T) *deviceEnrollmentHTTPFixture {
	t.Helper()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStore()
	t.Cleanup(func() {
		assignmentStore.Close()
	})
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:create", "ai-agent:device:read"}, "user-1"),
	}).Handler()
	return &deviceEnrollmentHTTPFixture{
		t:            t,
		server:       server,
		aiAgentStore: aiAgentStore,
	}
}
