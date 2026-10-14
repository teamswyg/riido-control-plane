package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDynamoDBStreamRelayPublishesOutboxRecords(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 3, 4, 5, 0, time.UTC)
	recordJSON := marshalOutboxRecordForStreamTest(t, TaskEvent{
		Seq: 1, TaskID: "task-a", AssignmentID: "asn-000001", AgentID: "jykim1",
		Type: EventAssignmentQueued, State: AssignmentQueued, Message: "queued", At: fixedNow,
	})
	requests := make(chan capturedDynamoDBRequest, 3)
	server := newPublishingStreamRelayServer(t, recordJSON, requests)
	defer server.Close()

	publisher := &fakeStreamRelayPublisher{}
	checkpoints := &fakeStreamRelayCheckpointStore{
		checkpoint: StreamRelayCheckpoint{
			SchemaVersion: StreamRelayCheckpointSchemaVersion,
			StreamARN:     streamRelayTestARN, ShardID: "shard-1",
			SequenceNumber: "41", UpdatedAt: fixedNow.Add(-time.Minute),
		},
		ok: true,
	}
	stats, err := RunDynamoDBStreamRelayOnce(context.Background(), DynamoDBStreamRelayConfig{
		Region: "ap-northeast-2", StreamARN: streamRelayTestARN, Endpoint: server.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKIDEXAMPLE", "SESSION"),
		Publisher:           publisher,
		CheckpointStore:     checkpoints,
		Now:                 func() time.Time { return fixedNow },
		ShardIteratorType:   "TRIM_HORIZON",
		Limit:               25,
	})
	if err != nil {
		t.Fatalf("RunDynamoDBStreamRelayOnce: %v", err)
	}

	assertPublishedStreamRelayStats(t, stats)
	if len(publisher.events) != 1 {
		t.Fatalf("events = %+v", publisher.events)
	}
	assertPublishedStreamRelayEvent(t, publisher.events[0])
	assertStreamRelayRequest(t, <-requests, dynamoDBStreamDescribeTarget)
	assertStreamRelayIteratorRequest(t, <-requests)
	assertStreamRelayRequest(t, <-requests, dynamoDBStreamGetRecordsTarget)
	if len(checkpoints.saved) != 1 || checkpoints.saved[0].SequenceNumber != "42" {
		t.Fatalf("saved checkpoints = %+v", checkpoints.saved)
	}
}
