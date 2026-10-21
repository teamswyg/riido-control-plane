package riidoaiserver

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestRecoveryJournalDeduplicatesAcrossReplicasAndOldRetries(t *testing.T) {
	a := Assignment{ID: "replica-asn", TaskID: "task", AgentID: "agent", RuntimeProvider: "codex", State: AssignmentLeased, LeaseToken: "lease", UpdatedAt: time.Now()}
	ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: a, LastEventSeq: 10}}
	first := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer first.Close()
	second := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer second.Close()
	req := recoveryEventRequest(a)
	original, err := first.RecordAgentEvent(context.Background(), a.AgentID, req)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := second.RecordAgentEvent(context.Background(), a.AgentID, req)
	if err != nil || !reflect.DeepEqual(original.Event, replay.Event) || len(ops.records) != 1 {
		t.Fatalf("replica replay changed committed event: %v records=%d", err, len(ops.records))
	}
	next := recoveryEventRequest(a)
	next.Metadata[metadatakeys.AssignmentEventKey.String()] = "recovery-check-2"
	next.Metadata[assignment.RecoveryAttemptKey] = "2"
	if _, err := second.RecordAgentEvent(context.Background(), a.AgentID, next); err != nil {
		t.Fatal(err)
	}
	replay, err = first.RecordAgentEvent(context.Background(), a.AgentID, req)
	if err != nil || !reflect.DeepEqual(original.Event, replay.Event) || len(ops.records) != 2 {
		t.Fatal("delayed retry was appended or changed its original ACK")
	}
	stale := recoveryEventRequest(a)
	stale.Metadata[metadatakeys.AssignmentEventKey.String()] = "unseen-old-phase"
	if _, err := first.RecordAgentEvent(context.Background(), a.AgentID, stale); err == nil {
		t.Fatal("older attempt moved the durable recovery watermark backwards")
	}
	req.Metadata[assignment.RecoveryReasonKey] = "provider_unavailable"
	if _, err := second.RecordAgentEvent(context.Background(), a.AgentID, req); err == nil {
		t.Fatal("same event key accepted a different transition")
	}
}
