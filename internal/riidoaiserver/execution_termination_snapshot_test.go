package riidoaiserver

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutionTerminationSnapshotRestartAndFailure(t *testing.T) {
	snapshot, err := NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	failing := &failingClaimSnapshot{SnapshotStore: snapshot}
	f := newTerminationHTTPFixture(t, StoreConfig{SnapshotStore: failing})
	failing.fail = true
	if w := serveTermination(f.handler, f.request(terminationFixtureBody)); w.Code != 503 {
		t.Fatalf("failed save status=%d", w.Code)
	}
	failing.fail = false
	body := strings.Replace(terminationFixtureBody, `:2}`, `:3}`, 1)
	first := serveTermination(f.handler, f.request(body))
	if first.Code != 200 {
		t.Fatalf("retry status=%d", first.Code)
	}
	reopened, err := OpenStoreWithConfig(t.Context(), StoreConfig{SnapshotStore: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	h := NewServer(ServerConfig{Assignment: reopened, AIAgentClient: f.devices}).Handler()
	again := serveTermination(h, f.request(body))
	if again.Code != 200 || again.Body.String() != first.Body.String() {
		t.Fatal("restart lost original receipt")
	}
	if w := serveTermination(h, f.request(terminationFixtureBody)); w.Code != 409 {
		t.Fatalf("restart changed receipt status=%d", w.Code)
	}
}

func TestExecutionTerminationUnsupportedJournalFailsClosed(t *testing.T) {
	journal := &runtimeFakeAssignmentOperationStore{}
	f := newTerminationHTTPFixture(t, StoreConfig{OperationStore: journal})
	before := len(journal.records)
	w := serveTermination(f.handler, f.request(terminationFixtureBody))
	if w.Code != 503 || len(journal.records) != before {
		t.Fatal("unsupported journal acknowledged or wrote receipt")
	}
}
