package riidoaiserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestExecutionClaimSurvivesSnapshotAndJournalRestart(t *testing.T) {
	for _, mode := range []string{"file snapshot", "journal"} {
		t.Run(mode, func(t *testing.T) {
			now := time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC)
			cfg := StoreConfig{Now: func() time.Time { return now }}
			var journal runtimeFakeAssignmentOperationStore
			if mode == "file snapshot" {
				var err error
				cfg.SnapshotStore, err = NewFileStoreSnapshot(filepath.Join(t.TempDir(), "state.json"))
				if err != nil {
					t.Fatal(err)
				}
			} else {
				cfg.OperationStore = &journal
			}
			actor := NewStoreWithConfig(cfg)
			a, err := actor.AssignTask(context.Background(), "task", AssignRequest{ComponentID: "component", AgentID: "agent", RuntimeProvider: "codex", Prompt: "fixture", CreatedBy: "requester"})
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(context.Background(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{deviceID: "device", ownerPrincipalID: "owner"})
			poll, err := actor.PollAgent(ctx, "agent", PollRequest{})
			if err != nil || poll.Action != PollStart {
				t.Fatalf("poll: %v", err)
			}
			actor.Close()
			if mode == "journal" {
				data, err := json.Marshal(journal.records)
				if err != nil {
					t.Fatal(err)
				}
				journal.records = nil
				if err := json.Unmarshal(data, &journal.records); err != nil {
					t.Fatal(err)
				}
			}
			reopened, err := OpenStoreWithConfig(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			observed, err := reopened.observeDeviceExecutions(context.Background(), "device")
			if err != nil || observed != deviceExecutionBlocked {
				t.Fatalf("restart lost non-owner claim: %s %v", observed, err)
			}
			if _, err := reopened.RecordAgentEvent(context.Background(), "agent", AgentEventRequest{AssignmentID: a.ID, State: AssignmentFailed, EventType: EventAssignmentFailed, Message: "fixture terminal"}); err != nil {
				t.Fatal(err)
			}
			reopened.Close()
			afterTerminal, err := OpenStoreWithConfig(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer afterTerminal.Close()
			observed, err = afterTerminal.observeDeviceExecutions(context.Background(), "device")
			if err != nil || observed != deviceExecutionUnknown {
				t.Fatalf("unproved terminal became safe: %s %v", observed, err)
			}
			assertExecutionClaimRestartHistory(t, cfg, journal.records)
		})
	}
}
