package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func verifySnapshotCase(tc caseSpec) (caseEvidence, error) {
	ctx := context.Background()
	path := filepath.Join(os.TempDir(), "riido-snapshot-"+tc.Name+".json")
	_ = os.Remove(path)
	snapshot, err := riidoaiserver.NewFileStoreSnapshot(path)
	if err != nil {
		return caseEvidence{}, err
	}
	store, err := riidoaiserver.OpenStoreWithConfig(ctx, riidoaiserver.StoreConfig{SnapshotStore: snapshot})
	if err != nil {
		return caseEvidence{}, err
	}
	assignment, err := assignAndPoll(ctx, store)
	if err != nil {
		return caseEvidence{}, err
	}
	if err := recordReady(ctx, store, assignment); err != nil {
		return caseEvidence{}, err
	}
	store.Close()
	return verifyReloadedSnapshot(ctx, path, tc)
}

func recordReady(ctx context.Context, store *riidoaiserver.Store, assignment riidoaiserver.Assignment) error {
	_, err := store.RecordAgentEvent(ctx, "agent-1", riidoaiserver.AgentEventRequest{
		AssignmentID: assignment.ID, TaskID: "task-a", DaemonID: "daemon-1",
		RuntimeID: "runtime-1", State: riidoaiserver.AssignmentReady,
		EventType: riidoaiserver.EventAssignmentReady, Message: "ready",
	})
	return err
}

func verifyReloadedSnapshot(ctx context.Context, path string, tc caseSpec) (caseEvidence, error) {
	snapshot, err := riidoaiserver.NewFileStoreSnapshot(path)
	if err != nil {
		return caseEvidence{}, err
	}
	store, err := riidoaiserver.OpenStoreWithConfig(ctx, riidoaiserver.StoreConfig{SnapshotStore: snapshot})
	if err != nil {
		return caseEvidence{}, err
	}
	defer store.Close()
	history, _, cancel, err := store.SubscribeTask(ctx, "task-a")
	if err != nil {
		return caseEvidence{}, err
	}
	cancel()
	return snapshotEvidence(ctx, store, tc, len(history))
}

func snapshotEvidence(ctx context.Context, store *riidoaiserver.Store, tc caseSpec, history int) (caseEvidence, error) {
	poll, err := store.PollAgent(ctx, "agent-1", pollRequest())
	if err != nil {
		return caseEvidence{}, err
	}
	metrics, err := store.Metrics(ctx)
	if err != nil {
		return caseEvidence{}, err
	}
	state := string(poll.Assignment.State)
	result := reloadedResult(tc, history, state, int(metrics.TaskEventsTotal))
	return result, verifySnapshotResult(tc, result)
}
