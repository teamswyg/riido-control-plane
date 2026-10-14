package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func assertPublishedStreamRelayStats(t *testing.T, stats DynamoDBStreamRelayStats) {
	t.Helper()
	if stats.ShardsDiscovered != 1 ||
		stats.BatchesRead != 1 ||
		stats.RecordsRead != 1 ||
		stats.RecordsPublished != 1 ||
		stats.RecordsSkipped != 0 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.CheckpointsLoaded != 1 || stats.CheckpointsSaved != 1 {
		t.Fatalf("checkpoint stats = %+v", stats)
	}
}

func assertPublishedStreamRelayEvent(t *testing.T, event StreamRelayEvent) {
	t.Helper()
	if event.SchemaVersion != StreamRelayEventSchemaVersion ||
		event.ShardID != "shard-1" ||
		event.SequenceNumber != "42" {
		t.Fatalf("relay event = %+v", event)
	}
	if event.EventID != "event-1" || event.EventName != "INSERT" {
		t.Fatalf("event metadata = %+v", event)
	}
	if event.Record.Event.TaskID != "task-a" ||
		event.Record.Event.Type != EventAssignmentQueued {
		t.Fatalf("outbox record = %+v", event.Record)
	}
}

func assertStreamRelayIteratorRequest(t *testing.T, request capturedDynamoDBRequest) {
	t.Helper()
	assertStreamRelayRequest(t, request, dynamoDBStreamGetShardIteratorTarget)
	var payload struct {
		ShardIteratorType string `json:"ShardIteratorType"`
		SequenceNumber    string `json:"SequenceNumber"`
	}
	if err := json.Unmarshal(request.body, &payload); err != nil {
		t.Fatalf("decode iterator payload: %v", err)
	}
	if payload.ShardIteratorType != "AFTER_SEQUENCE_NUMBER" ||
		payload.SequenceNumber != "41" {
		t.Fatalf("iterator payload = %+v", payload)
	}
}
