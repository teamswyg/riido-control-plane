package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type failingSink struct{}

func (failingSink) AppendTaskEvent(context.Context, riidoaiserver.TaskEvent) error {
	return errors.New("outbox unavailable")
}

func (failingSink) Close() error {
	return nil
}

func verifyOutboxFailureCase(tc caseSpec) (caseEvidence, error) {
	ctx := context.Background()
	store := riidoaiserver.NewStoreWithConfig(riidoaiserver.StoreConfig{Outbox: failingSink{}})
	defer store.Close()
	if _, err := store.AssignTask(ctx, "task-a", riidoaiserver.AssignRequest{
		ComponentID: "component-1", AgentID: "agent-1",
		RuntimeProvider: "codex", Prompt: "make hello world",
	}); err != nil {
		return caseEvidence{}, err
	}
	metrics, err := store.Metrics(ctx)
	if err != nil {
		return caseEvidence{}, err
	}
	result := caseEvidence{Name: tc.Name, Kind: tc.Kind, OutboxErrors: metrics.OutboxErrorsTotal, LatencySamples: metrics.EventAppendLatencySamplesTotal}
	if result.OutboxErrors != tc.WantOutboxErrors || result.LatencySamples != tc.WantLatencySamples {
		return result, fmt.Errorf("%s result=%+v", tc.Name, result)
	}
	return result, nil
}
