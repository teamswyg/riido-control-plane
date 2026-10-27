package riidoaiserver

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAssignmentContextQuotaCountsEveryAuthenticatedAttempt(t *testing.T) {
	for _, tc := range []struct {
		name, view, target string
		status             int
	}{
		{"snapshot", AssignmentContextViewBase, "", http.StatusOK},
		{"live", AssignmentContextViewAgents, "", http.StatusOK},
		{"denied", AssignmentContextViewTask, "outside-context", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
			// The fixture deliberately reuses request_id=cli_1. Replay is still an
			// attempt; otherwise live views could bypass the cap indefinitely.
			for i := range 100 {
				if code, body := f.read(tc.view, tc.target, f.leased.LeaseToken); code != tc.status {
					t.Fatalf("attempt %d: status=%d body=%s", i+1, code, body)
				}
			}
			if code, body := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusTooManyRequests {
				t.Fatalf("attempt 101: status=%d body=%s", code, body)
			}
		})
	}
}

func TestAssignmentContextQuotaRejectsConcurrentAttempt101(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	var allowed, denied, unexpected atomic.Int32
	var wg sync.WaitGroup
	for range 150 {
		wg.Go(func() {
			switch code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code {
			case http.StatusOK:
				allowed.Add(1)
			case http.StatusTooManyRequests:
				denied.Add(1)
			default:
				unexpected.Add(1)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 100 || denied.Load() != 50 || unexpected.Load() != 0 {
		t.Fatalf("allowed=%d quota_denied=%d unexpected=%d", allowed.Load(), denied.Load(), unexpected.Load())
	}
}

func TestAssignmentContextQuotaIgnoresInvalidLeaseAndPromptReads(t *testing.T) {
	f := newAssignmentContextFixture(t, NewMemoryAssignmentContextStore())
	for range 101 {
		if code, _ := f.read(AssignmentContextViewBase, "", "wrong-lease"); code != http.StatusForbidden {
			t.Fatalf("wrong lease status=%d", code)
		}
	}
	for i := range 100 {
		if code, body := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusOK {
			t.Fatalf("valid attempt %d: status=%d body=%s", i+1, code, body)
		}
	}
	if code, _ := f.read(AssignmentContextViewBase, "", f.leased.LeaseToken); code != http.StatusTooManyRequests {
		t.Fatalf("attempt 101: status=%d", code)
	}
}
