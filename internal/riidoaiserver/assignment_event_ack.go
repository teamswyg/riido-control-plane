package riidoaiserver

import (
	"time"
)

// Only finite lifecycle acknowledgements are retained here. Streaming progress
// uses its existing sequence dedupe; the independent recovery journal is intact.
const maxAssignmentEventACKReceipts = 8

type assignmentEventACKReceipt struct {
	KeyHash             string          `json:"key_hash"`
	PayloadHash         string          `json:"payload_hash"`
	Seq                 int64           `json:"seq"`
	At                  time.Time       `json:"at"`
	State               AssignmentState `json:"state"`
	Type                string          `json:"type"`
	ProviderSessionID   string          `json:"provider_session_id,omitempty"`
	ConfirmationReason  string          `json:"confirmation_reason,omitempty"`
	AssignmentUpdatedAt time.Time       `json:"assignment_updated_at"`
}
