package riidoaiserver

import (
	"errors"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

// Only finite lifecycle acknowledgements are retained here. Streaming progress
// uses its existing sequence dedupe; the independent recovery journal is intact.
func retainedAssignmentEventACK(req AgentEventRequest) bool {
	if strings.TrimSpace(req.Metadata[metadatakeys.AssignmentEventKey.String()]) == "" {
		return false
	}
	return req.EventType == EventAssignmentReady || (req.EventType != EventRiidoLog && isTerminal(req.State))
}

func newAssignmentEventACKReceipt(current Assignment, req AgentEventRequest, event TaskEvent, payloadHash string) *assignmentEventACKReceipt {
	if !retainedAssignmentEventACK(req) {
		return nil
	}
	return &assignmentEventACKReceipt{KeyHash: recoveryHash(req.Metadata[metadatakeys.AssignmentEventKey.String()]), PayloadHash: payloadHash, Seq: event.Seq, At: event.At, State: event.State, Type: event.Type, ProviderSessionID: current.ProviderSessionID, ConfirmationReason: string(current.ConfirmationReason), AssignmentUpdatedAt: current.UpdatedAt}
}

func assignmentEventACKsForRecord(record AssignmentOperationRecord) ([]assignmentEventACKReceipt, error) {
	var receipts []assignmentEventACKReceipt
	if record.ExpectedProjection != nil {
		receipts = append(receipts, record.ExpectedProjection.EventACKs...)
	}
	if record.eventACK == nil {
		return receipts, nil
	}
	for _, existing := range receipts {
		if existing.KeyHash == record.eventACK.KeyHash {
			if existing.PayloadHash != record.eventACK.PayloadHash {
				return nil, errors.New("assignment event acknowledgement conflict")
			}
			return receipts, nil
		}
	}
	if len(receipts) >= maxAssignmentEventACKReceipts {
		return nil, errors.New("assignment event acknowledgement budget exhausted")
	}
	return append(receipts, *record.eventACK), nil
}
