package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func callAuthorizedMetrics(m manifest) (int, riidoaiserver.MetricsSnapshot, map[string]bool, error) {
	store, err := seededStore()
	if err != nil {
		return 0, riidoaiserver.MetricsSnapshot{}, nil, err
	}
	defer store.Close()
	metrics := riidoaiserver.NewHTTPTransactionMetrics()
	metrics.ObserveHTTPTransaction(riidoaiserver.HTTPTransactionObservation{
		Method: http.MethodPost, Route: "/v1/agents/{agent_id}/poll", StatusCode: http.StatusAccepted,
	})
	server, err := metricsServer(store, metrics, scopedAuthorizer("metrics-token", "metrics:read"))
	if err != nil {
		return 0, riidoaiserver.MetricsSnapshot{}, nil, err
	}
	status, body := callMetrics(server, m, "metrics-token")
	snapshot, fields, err := decodeSnapshot(body)
	if err != nil {
		return 0, riidoaiserver.MetricsSnapshot{}, nil, err
	}
	if snapshot.SchemaVersion == "" {
		return 0, snapshot, fields, errors.New("missing metrics schema version")
	}
	return status, snapshot, fields, nil
}

func seededStore() (*riidoaiserver.Store, error) {
	store := riidoaiserver.NewStore()
	assignment, err := store.AssignTask(context.Background(), "task-a", riidoaiserver.AssignRequest{
		ComponentID: "component-a", AgentID: "agent-a", RuntimeProvider: "codex", Prompt: "ship it",
	})
	if err != nil {
		return nil, err
	}
	if _, err := store.PollAgent(context.Background(), "agent-a", riidoaiserver.PollRequest{DaemonID: "daemon-a"}); err != nil {
		return nil, err
	}
	_, err = store.RecordAgentEvent(context.Background(), "agent-a", riidoaiserver.AgentEventRequest{
		AssignmentID: assignment.ID, DaemonID: "daemon-a", State: riidoaiserver.AssignmentRunning,
		EventType: riidoaiserver.EventAssignmentRunning,
	})
	return store, err
}
