package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
)

type daemonBindingsRuntimeStore struct {
	*DevelopmentAIAgentClientStore
	response AgentRuntimeBindingListResponse
	err      error
	gotID    string
}

func (s *daemonBindingsRuntimeStore) ListAIAgentDaemonAgentBindings(_ context.Context, _ AuthorizationResult, deviceID string) (AgentRuntimeBindingListResponse, error) {
	s.gotID = deviceID
	return s.response, s.err
}

type daemonBindingsCredentialStore struct {
	principal AuthorizationResult
	err       error
}

func (s daemonBindingsCredentialStore) EnrollDeviceCredential(context.Context, AuthorizationResult, string, EnrollDeviceRequest) (EnrollDeviceResponse, error) {
	return EnrollDeviceResponse{}, nil
}

func (s daemonBindingsCredentialStore) AuthorizeDeviceCredential(context.Context, string, string, AuthorizationRequest) (AuthorizationResult, error) {
	if s.err != nil {
		return AuthorizationResult{}, s.err
	}
	return s.principal, nil
}

func serveDaemonBindings(handler http.Handler, deviceID, secret string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/daemon/agent-bindings", nil)
	if deviceID != "" {
		req.Header.Set(deviceIDHeader, deviceID)
	}
	if secret != "" {
		req.Header.Set(deviceSecretHeader, secret)
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func newDaemonBindingsTestServer(runtimeStore *daemonBindingsRuntimeStore) http.Handler {
	if runtimeStore.DevelopmentAIAgentClientStore == nil {
		runtimeStore.DevelopmentAIAgentClientStore = NewDevelopmentAIAgentClientStore()
	}
	credential := daemonBindingsCredentialStore{
		principal: AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-1"},
	}
	return NewServer(ServerConfig{AIAgentClient: runtimeStore, DeviceCredentials: credential}).Handler()
}
