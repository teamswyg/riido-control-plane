package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

type bootstrapDevicesErrorStore struct {
	*DevelopmentAIAgentClientStore
	bootstrapErr error
	devicesErr   error
	reconcileErr error
}

func (s bootstrapDevicesErrorStore) BootstrapAIAgentClient(ctx context.Context, principal AuthorizationResult, clientKind ClientKind) (ClientBootstrapResponse, error) {
	if s.bootstrapErr != nil {
		return ClientBootstrapResponse{}, s.bootstrapErr
	}
	return s.DevelopmentAIAgentClientStore.BootstrapAIAgentClient(ctx, principal, clientKind)
}

func (s bootstrapDevicesErrorStore) ListAIAgentDevices(ctx context.Context, principal AuthorizationResult) (DeviceRuntimeListResponse, error) {
	if s.devicesErr != nil {
		return DeviceRuntimeListResponse{}, s.devicesErr
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentDevices(ctx, principal)
}

func (s bootstrapDevicesErrorStore) ReconcileAIAgentActiveThreadProjections(context.Context, AuthorizationResult, string, AssignmentProjectionReader) (bool, error) {
	return false, s.reconcileErr
}

func bootstrapDevicesErrorServer(t *testing.T, store AIAgentClientStore, assignment AssignmentStore) http.Handler {
	t.Helper()
	return newTaskThreadReadErrorTestServer(t, []string{"ai-agent:*"}, store, assignment)
}
