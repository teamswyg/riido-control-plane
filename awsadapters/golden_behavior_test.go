package awsadapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"
)

const awsAdaptersGoldenHash = "4b2edc9037b7cf27bc857154fd5e2d92bebc13fcc302351b83ec847f5c2500ec"

func TestAWSAdaptersFacadeBehaviorGolden(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	credentials, err := provider.Credentials(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	body := canonicalAWSAdapterGolden(t, credentials)
	if got := sha256Hex(body); got != awsAdaptersGoldenHash {
		t.Fatalf("awsadapters facade golden drifted: %s\n%s", got, body)
	}
}

func canonicalAWSAdapterGolden(t *testing.T, credentials AWSCredentials) []byte {
	t.Helper()
	body, err := json.MarshalIndent(struct {
		Credentials AWSCredentials   `json:"credentials"`
		Event       StreamRelayEvent `json:"event"`
	}{
		Credentials: credentials,
		Event: StreamRelayEvent{
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
		},
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
