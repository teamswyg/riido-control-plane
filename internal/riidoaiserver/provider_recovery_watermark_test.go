package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestRecoveryWatermarkRoundtripBoundAndPublicIsolation(t *testing.T) {
	a := Assignment{ID: "asn", TaskID: "task", AgentID: "agent", State: AssignmentLeased}
	record := AssignmentOperationRecord{AssignmentID: a.ID, Assignment: a}
	// A deliberately larger sequence than the three-attempt/one-switch engine
	// can emit proves the hard upper bound instead of relying on caller budgets.
	for i := range maxRecoveryReceipts {
		metadata := (assignment.RecoveryContext{State: "checking", Provider: "codex", Attempt: 1}).Metadata()
		metadata[metadatakeys.AssignmentEventKey.String()] = fmt.Sprintf("private-key-%d", i)
		record.Events = []TaskEvent{{Type: EventRiidoLog, Seq: int64(i + 1), At: time.Now(), State: AssignmentLeased, Metadata: metadata}}
		item, err := assignmentProjectionDynamoDBItem(record)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := assignmentProjectionFromDynamoDBItem(item)
		if err != nil || decoded.Recovery == nil || len(decoded.Recovery.Receipts) != i+1 {
			t.Fatalf("roundtrip receipts=%d: %v", i, err)
		}
		if strings.Contains(item["recovery_watermark_json"]["S"], "private-key") {
			t.Fatal("projection copied raw event keys")
		}
		projection := AssignmentProjection(decoded)
		record.ExpectedProjection = &projection
	}
	if _, err := assignmentProjectionDynamoDBItem(record); err == nil {
		t.Fatal("journal exceeded bounded receipt budget")
	}
	record.Events = []TaskEvent{{Type: EventAssignmentCompleted, Seq: 17}}
	item, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := assignmentProjectionFromDynamoDBItem(item)
	if err != nil || len(decoded.Recovery.Receipts) != maxRecoveryReceipts {
		t.Fatal("normal terminal event discarded the recovery receipts")
	}
	public, _ := json.Marshal(AssignmentProjection(decoded))
	if strings.Contains(string(public), "receipt") || strings.Contains(string(public), "watermark") {
		t.Fatal("internal journal leaked into assignment wire payload")
	}
}
