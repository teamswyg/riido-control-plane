package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type dashboardRevisionFixture struct {
	memoryAIAgentClientSnapshotStore
	checks int
}

func (d *dashboardRevisionFixture) dashboardSnapshotVersion(context.Context) (int64, error) {
	d.checks++
	return d.snapshot.Version, nil
}

type dashboardEventRecorder struct {
	*httptest.ResponseRecorder
	visible chan struct{}
}

func (r dashboardEventRecorder) Write(p []byte) (int, error) {
	n, err := r.ResponseRecorder.Write(p)
	if strings.Contains(string(p), "agent_work_status_changed") {
		select {
		case r.visible <- struct{}{}:
		default:
		}
	}
	return n, err
}

func TestDashboardIdleReplicaRefreshesAtKeepalive(t *testing.T) {
	const agent = "agent-owned-codex"
	writer := NewDevelopmentAIAgentClientStore()
	snapshot, err := writer.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Version = 1
	db := &dashboardRevisionFixture{memoryAIAgentClientSnapshotStore: memoryAIAgentClientSnapshotStore{snapshot: snapshot, ok: true}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, err := OpenPersistentAIAgentClientStore(ctx, NewDevelopmentAIAgentClientStore(), db)
	if err != nil {
		t.Fatal(err)
	}
	_, live, stop, err := store.SubscribeAIAgentClientEvents(context.WithValue(ctx, dashboardStreamScope{}, agent), AuthorizationResult{WorkspaceID: defaultAIAgentClientWorkspaceID, PrincipalID: "user-1"})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	writer.mu.Lock()
	writer.appendClientEventLocked(AgentClientEventWorkStatusChanged, AgentWorkStatusChangedEvent{EventType: AgentClientEventWorkStatusChanged, AgentID: agent, AssignmentID: "remote", TaskID: "visible", AssignmentState: AgentAssignmentStateQueued})
	writer.mu.Unlock()
	snap, err := writer.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	snap.Version = db.snapshot.Version + 1
	db.snapshot = snap
	server := NewServer(ServerConfig{AIAgentClient: store, Authorizer: dashboardTestAuthorizer{}, TaskContext: &dashboardTestContext{}})
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer viewer-token")
	rec := dashboardEventRecorder{httptest.NewRecorder(), make(chan struct{}, 1)}
	ticks := make(chan time.Time, 1)
	ticks <- time.Now()
	done := make(chan struct{})
	go func() {
		defer close(done)
		server.deliverDashboardEvents(rec, req, live, defaultAIAgentClientWorkspaceID, agent, ticks)
	}()
	select {
	case <-rec.visible:
	case <-time.After(2 * time.Second):
		cancel()
		<-done
		t.Fatalf("keepalive did not deliver remote queue: checks=%d loads=%d localVersion=%d sharedVersion=%d events=%d subscribers=%d body=%s", db.checks, db.loads, store.version, db.snapshot.Version, len(store.events), len(store.subscribers), rec.Body.String())
	}
	cancel()
	<-done
	reads := db.loads
	if err := store.refreshDashboardSubscription(context.Background()); err != nil {
		t.Fatal(err)
	}
	if db.loads != reads || db.checks != 1 {
		t.Fatal("subscribers did not share revision check")
	}
	store.dashboardLastRefresh = time.Time{}
	if err := store.refreshDashboardSubscription(context.Background()); err != nil {
		t.Fatal(err)
	}
	if db.checks != 2 || db.loads != reads {
		t.Fatal("unchanged revision replayed snapshot")
	}
}
