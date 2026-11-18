package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type clientEventsFailingSnapshotStore struct {
	memoryAIAgentClientSnapshotStore
	loadErr error
}

func (s *clientEventsFailingSnapshotStore) LoadAIAgentClientSnapshot(ctx context.Context) (AIAgentClientSnapshot, bool, error) {
	if s.loadErr != nil {
		s.loads++
		return AIAgentClientSnapshot{}, false, s.loadErr
	}
	return s.memoryAIAgentClientSnapshotStore.LoadAIAgentClientSnapshot(ctx)
}

// Keep the persistent reader but hide its optional live subscription interface.
type clientEventsPersistentReaderOnlyStore struct{ AIAgentClientStore }

func TestHTTPAIAgentClientEventsSnapshotFailure(t *testing.T) {
	for _, failure := range []string{"load", "restore"} {
		for _, mode := range []string{"subscriber", "reader", "notifications"} {
			for _, path := range []string{"/v1/client/ai-agent/events", "/v2/client/workspaces/workspace-dev-riid/ai-agent/events"} {
				t.Run(failure+"/"+mode+path, func(t *testing.T) {
					snapshots := &clientEventsFailingSnapshotStore{}
					persistent, err := OpenPersistentAIAgentClientStore(t.Context(), nil, snapshots)
					if err != nil {
						t.Fatal(err)
					}
					persistent.snapshotReloadInterval = 0
					const privateDetail = "private snapshot storage detail"
					if failure == "load" {
						snapshots.loadErr = errors.New(privateDetail)
					} else {
						snapshots.snapshot.SchemaVersion = privateDetail
					}
					var store AIAgentClientStore = persistent
					if mode == "reader" {
						store = clientEventsPersistentReaderOnlyStore{AIAgentClientStore: persistent}
					}
					if mode == "notifications" {
						path += "?notifications=1&workspace_id=workspace-dev-riid"
					}
					server := newClientEventsErrorTestServer(t, nil, store)
					request := httptest.NewRequest(http.MethodGet, path, nil)
					request.Header.Set("Authorization", "Bearer user-token")
					response := httptest.NewRecorder()
					loads := snapshots.loads
					server.ServeHTTP(response, request)
					if snapshots.loads != loads+1 {
						t.Fatal("request did not reload the snapshot")
					}
					if response.Code != http.StatusServiceUnavailable {
						t.Fatalf("status=%d want=%d", response.Code, http.StatusServiceUnavailable)
					}
					if strings.Contains(response.Body.String(), privateDetail) {
						t.Fatal("response exposed private storage detail")
					}
					if len(persistent.subscribers) != 0 {
						t.Fatal("failed request retained a subscriber")
					}
				})
			}
		}
	}
}
