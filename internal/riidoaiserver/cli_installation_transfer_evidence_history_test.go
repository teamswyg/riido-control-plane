package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationTransferEvidencePreservesHistory(t *testing.T) {
	for _, state := range []string{"consumed", "expired", "retired"} {
		t.Run(state, func(t *testing.T) {
			f := newCLITransferFixture(t)
			if state == "consumed" {
				if status, _ := f.confirm(t); status != http.StatusOK {
					t.Fatalf("confirm status=%d", status)
				}
			} else {
				f.operations.now = func() time.Time { return f.intent.ExpiresAt.Add(time.Second) }
			}
			if state == "retired" {
				body := installationRetireBody(f.cli, f.intent.IntentID)
				if w := f.post(t, installationRetirePath, body, "service", ""); w.Code != http.StatusOK {
					t.Fatalf("retire status=%d", w.Code)
				}
				f.operations.now = func() time.Time { return f.intent.ExpiresAt.Add(-time.Second) }
			}
			assertTransferEvidenceRead(t, f, transferEvidenceBody(t, f), true)
			if state != "consumed" {
				if status, _ := f.confirm(t); status != http.StatusGone {
					t.Fatalf("history renewed authority: status=%d", status)
				}
				f.assertOriginal(t)
			}
		})
	}
}
