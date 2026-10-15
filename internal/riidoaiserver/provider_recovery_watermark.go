package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

// Three attempts, one switch: observed + 3*(backoff/checking/switching) +
// dispatched(started/confirmed) + one outcome fits sixteen receipts. This is
// private projection data; no prompt, raw event key, credential or public field.
const maxRecoveryReceipts = 16

type recoveryWatermark struct {
	Context  assignment.RecoveryContext `json:"context"`
	Receipts []recoveryReceipt          `json:"receipts"`
}

type recoveryReceipt struct {
	KeyHash            string                        `json:"key_hash"`
	ContextHash        string                        `json:"context_hash"`
	Seq                int64                         `json:"seq"`
	At                 time.Time                     `json:"at"`
	State              AssignmentState               `json:"state"`
	ConfirmationReason assignment.ConfirmationReason `json:"confirmation_reason,omitempty"`
}

func recoveryHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func recoveryContextHash(context assignment.RecoveryContext) string {
	c := context.Sanitized()
	return recoveryHash(strings.Join([]string{c.State, c.Reason, c.Provider, c.Circuit, c.OutcomeCertainty, strconv.Itoa(c.Attempt), strconv.Itoa(c.SwitchCount)}, "\x00"))
}

func recoveryWatermarkForRecord(record AssignmentOperationRecord) (*recoveryWatermark, error) {
	var current *recoveryWatermark
	if record.ExpectedProjection != nil && record.ExpectedProjection.Recovery != nil {
		cloned := *record.ExpectedProjection.Recovery
		cloned.Receipts = append([]recoveryReceipt(nil), cloned.Receipts...)
		current = &cloned
	}
	for _, event := range record.Events {
		context := assignment.RecoveryContextFromMetadata(event.Metadata)
		if event.Type != EventRiidoLog || context == nil {
			continue
		}
		if current == nil {
			current = &recoveryWatermark{}
		}
		if len(current.Receipts) >= maxRecoveryReceipts {
			return nil, errors.New("provider recovery journal budget exhausted")
		}
		current.Context = *context
		current.Receipts = append(current.Receipts, recoveryReceipt{KeyHash: recoveryHash(event.Metadata[metadatakeys.AssignmentEventKey.String()]), ContextHash: recoveryContextHash(*context), Seq: event.Seq, At: event.At, State: event.State, ConfirmationReason: event.ConfirmationReason})
	}
	return current, nil
}
