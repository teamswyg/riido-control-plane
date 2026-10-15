package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failingRuntimeSnapshotStore struct {
	*DevelopmentAIAgentClientStore
	failure error
}

func (s failingRuntimeSnapshotStore) SyncAIAgentDaemonRuntimeSnapshot(context.Context, AuthorizationResult, DeviceRuntimeSnapshotSyncRequest) (DeviceRuntimeSnapshotSyncResponse, error) {
	return DeviceRuntimeSnapshotSyncResponse{}, s.failure
}

func TestRuntimeSnapshotTypedFailureHTTPStatus(t *testing.T) {
	for _, tc := range []struct {
		name    string
		failure error
		status  int
	}{{"deadline", context.DeadlineExceeded, 504}, {"wrapped-deadline", errors.Join(errors.New("storage"), context.DeadlineExceeded), 504}, {"cancelled", context.Canceled, 408}, {"storage", errors.New("private upstream body"), 503}, {"validation", invalidRuntimeSnapshot(errors.New("invalid runtime")), 400}, {"forbidden", ErrAuthorizationForbidden, 403}} {
		t.Run(tc.name, func(t *testing.T) {
			auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-runtime-snapshot", Token: "user-token", Scopes: []string{"ai-agent:*"}}})
			if err != nil {
				t.Fatal(err)
			}
			handler := NewServer(ServerConfig{AIAgentClient: failingRuntimeSnapshotStore{DevelopmentAIAgentClientStore: NewDevelopmentAIAgentClientStore(), failure: tc.failure}, Authorizer: auth}).Handler()
			enrollment := enrollRuntimeSnapshotDevice(t, handler)
			req := runtimeSnapshotRequest(`{"daemon_id":"daemon","runtimes":[{"runtime_id":"runtime","kind":"codex"}]}`, enrollment)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tc.status || ackContainsAny(response.Body.String(), "private upstream body") {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestRuntimeSnapshotProductionValidationIsTyped(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	for _, req := range []DeviceRuntimeSnapshotSyncRequest{{DeviceID: "device"}, {DeviceID: "device", DaemonID: "daemon"}, {DeviceID: "device", DaemonID: "daemon", Runtimes: []RuntimeSnapshotRecord{{RuntimeID: "runtime"}}}} {
		_, err := store.SyncAIAgentDaemonRuntimeSnapshot(context.Background(), AuthorizationResult{PrincipalID: "owner"}, req)
		status, _ := runtimeSnapshotErrorResponse(err)
		if err == nil || status != http.StatusBadRequest {
			t.Fatalf("validation classification=%d %v", status, err)
		}
	}
}
