package riidoaiserver

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func recoveryEventRequest(a Assignment) AgentEventRequest {
	metadata := (assignment.RecoveryContext{State: "checking", Reason: "transient_preflight", Provider: "codex", Circuit: "closed", OutcomeCertainty: "not_started", Attempt: 1}).Metadata()
	metadata[metadatakeys.AssignmentEventKey.String()] = "recovery-check-1"
	metadata["lease_token"] = a.LeaseToken
	return AgentEventRequest{AssignmentID: a.ID, TaskID: a.TaskID, DaemonID: "daemon", DeviceID: "device", RuntimeID: "runtime", EventType: EventRiidoLog, Metadata: metadata, Message: "Bearer secret /Users/private/customer"}
}

func TestRecoveryEventFencedSanitizedPersistedAndObservedOnce(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	current := Assignment{ID: "recovery-asn", TaskID: "task", AgentID: "agent", RuntimeProvider: "codex", State: AssignmentLeased, LeaseToken: "private-lease", UpdatedAt: time.Now()}
	ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: current, LastEventSeq: 10}}
	recorder := &recordingTraceRecorder{}
	ctx := WithTraceRecorder(context.Background(), recorder)
	store := NewStoreWithConfig(StoreConfig{OperationStore: ops})
	defer store.Close()
	req := recoveryEventRequest(current)
	req.Metadata["untrusted_customer_path"] = "/Users/private/customer"
	first, err := store.RecordAgentEvent(ctx, "agent", req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.RecordAgentEvent(ctx, "agent", req)
	if err != nil {
		t.Fatal(err)
	}
	if first.Event.Seq != second.Event.Seq || len(ops.records) != 1 {
		t.Fatal("duplicate created another transition")
	}
	if first.Event.Metadata["lease_token"] != "" || first.Event.Metadata["untrusted_customer_path"] != "" || first.Event.Message != "" {
		t.Fatal("recovery journal leaked lease/output")
	}
	if req.Metadata["lease_token"] != "private-lease" {
		t.Fatal("caller metadata mutated")
	}
	if got := strings.Count(output.String(), "event=provider_recovery_transition"); got != 1 {
		t.Fatalf("observations=%d logs=%q", got, output.String())
	}
	spans := recorder.snapshot()
	count := 0
	for _, span := range spans {
		if span.Name == "provider recovery transition" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("recovery trace count=%d", count)
	}
	for _, forbidden := range []string{"private-lease", "Bearer secret", "/Users/private/customer"} {
		if strings.Contains(output.String(), forbidden) {
			t.Fatal("sensitive recovery log")
		}
	}
	ops.projection.Assignment.State = AssignmentCancelling
	if _, err := store.RecordAgentEvent(ctx, "agent", req); err == nil {
		t.Fatal("late checking replay can authorize execution after cancel")
	}
}
