package main

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	srv "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type scenario struct {
	name  string
	risk  string
	next  string
	build func(config) (func() error, error)
}

func scenarios() []scenario {
	return []scenario{
		{"http_endpoint_threads_v3", "v3 thread history endpoint covers auth, route metrics, reconcile, projection, and JSON response pressure", "track refresh pressure before changing v3 history projection internals", buildHTTPEndpointThreadsV3},
		{"http_metrics_observe", "route metrics lock/allocation pressure on endpoint hot paths", "watch lock cost before adding route dimensions", buildHTTPMetrics},
		{"store_metrics_observe", "store operation metrics lock/allocation pressure on DB hot paths", "keep operation vocabulary bounded and compare with live EMF", buildStoreMetrics},
		{"progress_ingest_fragment", "fine-grained provider fragments can amplify normalize, merge, event, and fanout cost", "measure fragment cost before changing provider batching", buildProgressIngest},
		{"thread_stream_subscription", "SSE filter computation can copy active thread state under task fanout", "prefer active target projection over full thread projection", buildThreadSubscription},
		{"client_event_subscriber_fanout", "workspace SSE subscribers add per-event visibility and channel-send work", "measure subscriber fanout before adding richer live events", buildClientSubscriberFanout},
		{"thread_history_v3", "v3 thread history can copy historical messages during refresh/recovery", "split immutable snapshots and cap per-refresh projection work", buildThreadHistory},
	}
}

func buildHTTPMetrics(config) (func() error, error) {
	metrics := srv.NewHTTPTransactionMetrics()
	obs := srv.HTTPTransactionObservation{
		Method: http.MethodGet, Route: "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads",
		StatusCode: http.StatusOK, Duration: 2 * time.Millisecond,
	}
	return func() error { metrics.ObserveHTTPTransaction(obs); return nil }, nil
}

func buildStoreMetrics(config) (func() error, error) {
	metrics := srv.NewStoreOperationMetrics()
	ops := []srv.StoreOperationName{srv.StoreOperationPollAssignment, srv.StoreOperationLeaseAssignment, srv.StoreOperationAppendEvent}
	var i atomic.Int64
	return func() error {
		next := int(i.Add(1))
		metrics.ObserveStoreOperation(srv.StoreOperationObservation{Operation: ops[next%len(ops)], Duration: time.Millisecond})
		return nil
	}, nil
}

func buildThreadSubscription(cfg config) (func() error, error) {
	store, principal, taskID, err := pressureFixture(cfg)
	if err != nil {
		return nil, err
	}
	return func() error {
		_, err := store.GetAIAgentTaskThreadStreamSubscription(context.Background(), principal, taskID)
		return err
	}, nil
}

func buildThreadHistory(cfg config) (func() error, error) {
	store, principal, taskID, err := pressureFixture(cfg)
	if err != nil {
		return nil, err
	}
	return func() error {
		_, err := store.ListAIAgentTaskThreadHistory(context.Background(), principal, taskID)
		return err
	}, nil
}
