package awsadapters

import (
	"context"
	"testing"
	"time"
)

type recordingPublisher struct {
	events []StreamRelayEvent
}

func (p *recordingPublisher) PublishStreamRelayEvent(_ context.Context, event StreamRelayEvent) error {
	p.events = append(p.events, event)
	return nil
}

func TestStaticCredentialsFacade(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "TOKEN")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	var _ AWSCredentialsProvider = provider
	credentials, err := provider.Credentials(context.Background())
	if err != nil {
		t.Fatalf("Credentials: %v", err)
	}
	if credentials.AccessKeyID != "AKID" ||
		credentials.SecretAccessKey != "SECRET" ||
		credentials.SessionToken != "TOKEN" {
		t.Fatalf("credentials = %+v", credentials)
	}
}

func TestStreamRelayFacadeTypesCompile(t *testing.T) {
	publisher := &recordingPublisher{}
	var _ StreamRelayPublisher = publisher

	event := StreamRelayEvent{
		SchemaVersion:  "riido-dynamodb-stream-relay-event.v1",
		StreamARN:      "arn:aws:dynamodb:ap-northeast-2:123456789012:table/events/stream/2026-05-27T00:00:00.000",
		ShardID:        "shard-0001",
		SequenceNumber: "1",
		EventID:        "event-1",
		EventName:      "INSERT",
		Record: OutboxRecord{
			SchemaVersion: OutboxRecordSchemaVersion,
			Event: TaskEvent{
				Seq:          1,
				TaskID:       "task-1",
				AssignmentID: "asn-1",
				AgentID:      "agent-1",
				Type:         EventAssignmentQueued,
				State:        AssignmentQueued,
				At:           time.Unix(0, 0).UTC(),
			},
		},
	}
	if err := publisher.PublishStreamRelayEvent(context.Background(), event); err != nil {
		t.Fatalf("PublishStreamRelayEvent: %v", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("events = %d, want 1", len(publisher.events))
	}
	if publisher.events[0].Record.Event.State != AssignmentQueued {
		t.Fatalf("state = %q", publisher.events[0].Record.Event.State)
	}
}
