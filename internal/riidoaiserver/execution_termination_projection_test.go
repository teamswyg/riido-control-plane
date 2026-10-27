package riidoaiserver

import (
	"path/filepath"
	"testing"
)

func TestExecutionTerminationProjectionMustBeTerminalInSameEpoch(t *testing.T) {
	for _, mode := range []string{"active", "missing", "new epoch", "active after receipt"} {
		t.Run(mode, func(t *testing.T) {
			snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
			if err != nil {
				t.Fatal(err)
			}
			f := newTerminationHTTPFixture(t, StoreConfig{SnapshotStore: snapshot})
			if mode == "active after receipt" {
				if w := serveTermination(f.handler, f.request(terminationFixtureBody)); w.Code != 200 {
					t.Fatalf("initial status=%d", w.Code)
				}
			}
			state, _, err := snapshot.LoadStoreSnapshot(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			want := 409
			if mode == "missing" {
				state.Assignments = nil
				state.AgentAssignments = map[string][]string{}
				state.Tasks = nil
				want = 403
			} else {
				state.Assignments[0].State = AssignmentRunning
				if mode == "new epoch" {
					state.Assignments[0].LeaseToken = "new-epoch"
					want = 200
				}
			}
			if err := snapshot.SaveStoreSnapshot(t.Context(), state); err != nil {
				t.Fatal(err)
			}
			reopened, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: snapshot})
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			h := NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices}).Handler()
			w := serveTermination(h, f.request(terminationFixtureBody))
			if w.Code != want {
				t.Fatalf("status=%d want=%d", w.Code, want)
			}
		})
	}
}
