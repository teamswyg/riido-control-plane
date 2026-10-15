package riidoaiserver

import (
	"testing"
	"time"
)

func TestRecoveryApprovalFingerprintEncodingFailsClosed(t *testing.T) {
	current := Assignment{ID: "invalid-time", CreatedAt: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)}
	if approval, err := newRecoveryCandidateApproval(current, "cursor"); err == nil || approval != nil {
		t.Fatal("invalid assignment encoding granted approval")
	}
	store := &Store{recoveryApproval: &recoveryCandidateApproval{provider: "cursor", expiresAt: time.Now().Add(time.Second)}}
	if err := store.checkRecoveryCandidateApproval(current, "cursor"); err == nil {
		t.Fatal("empty fingerprint was accepted")
	}
}
