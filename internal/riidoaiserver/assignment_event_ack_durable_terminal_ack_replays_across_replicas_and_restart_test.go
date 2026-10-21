package riidoaiserver

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestDurableTerminalACKReplaysAcrossReplicasAndRestart(t *testing.T) {
	for _, terminal := range []AssignmentState{AssignmentFailed, AssignmentCompleted, AssignmentCancelled} {
		t.Run(string(terminal), func(t *testing.T) {
			ops, req := lifecycleACKFixture(t, AssignmentRunning)
			req.State = terminal
			switch terminal {
			case AssignmentCompleted:
				req.EventType = EventAssignmentCompleted
			case AssignmentCancelled:
				req.EventType = EventAssignmentCancelled
			}
			ops.conflictsRemaining = 1
			first := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			second := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer first.Close()
			defer second.Close()
			ack, err := first.RecordAgentEvent(context.Background(), "agent", req)
			if err != nil {
				t.Fatal(err)
			}
			replay, err := second.RecordAgentEvent(context.Background(), "agent", req)
			if err != nil || !reflect.DeepEqual(ack, replay) || len(ops.records) != 1 {
				t.Fatalf("replica ACK changed: first=%+v replay=%+v err=%v records=%d", ack, replay, err, len(ops.records))
			}
			// Exercise actual private Dynamo projection serialization before restart.
			record := ops.records[0]
			item, err := assignmentProjectionDynamoDBItem(record)
			if err != nil {
				t.Fatal(err)
			}
			projection, err := assignmentProjectionFromDynamoDBItem(item)
			if err != nil {
				t.Fatal(err)
			}
			restartedOps := &conditionalTestOperations{projection: AssignmentProjection(projection)}
			restarted := NewStoreWithConfig(StoreConfig{OperationStore: restartedOps})
			defer restarted.Close()
			restored, err := restarted.RecordAgentEvent(context.Background(), "agent", req)
			if err != nil || !reflect.DeepEqual(ack, restored) || len(restartedOps.records) != 0 {
				t.Fatalf("restart replay=%+v err=%v", restored, err)
			}
			encoded, _ := json.Marshal(projection.Assignment)
			if ackContainsAny(string(encoded), "key_hash", "payload_hash", "event_ack_receipts") {
				t.Fatal("private receipt leaked in public assignment")
			}
		})
	}
}
