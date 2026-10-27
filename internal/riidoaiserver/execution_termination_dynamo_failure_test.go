package riidoaiserver

import (
	"net/http"
	"testing"
)

func TestExecutionTerminationDynamoPartialFailureIsRetryable(t *testing.T) {
	for _, mode := range []string{"denied", "lost ack", "active race", "claim race", "missing projection", "epoch race", "legacy epoch"} {
		t.Run(mode, func(t *testing.T) {
			f, newStore, a := newTerminationDynamoFixture(t)
			h := terminationDynamoHTTP(t, newStore(), a)
			want := 200
			projectionKey := dynamoDBAssignmentProjectionPK(a.ID) + "|" + dynamoDBAssignmentProjectionSK
			switch mode {
			case "denied":
				f.deny = true
				want = 503
			case "lost ack":
				f.loseACK = true
			case "active race":
				f.beforeCommit = func(items map[string]map[string]map[string]string) {
					active := a
					active.State = AssignmentRunning
					items[projectionKey] = terminationProjectionItem(t, active, 4)
				}
				want = 409
			case "claim race":
				f.beforeCommit = func(items map[string]map[string]map[string]string) {
					for key, item := range items {
						if item["schema"]["S"] == executionClaimSchema {
							delete(items, key)
						}
					}
				}
				want = 403
			case "epoch race":
				active := a
				active.State = AssignmentRunning
				active.LeaseToken = "new-epoch"
				f.items[projectionKey] = terminationProjectionItem(t, active, 3)
				f.beforeCommit = func(items map[string]map[string]map[string]string) {
					active.LeaseToken = a.LeaseToken
					items[projectionKey] = terminationProjectionItem(t, active, 3)
				}
				want = 409
			case "legacy epoch":
				delete(f.items[projectionKey], "lease_epoch")
				want = 403
			case "missing projection":
				delete(f.items, projectionKey)
				want = 403
			}
			w := serveTermination(h.handler, h.request(terminationFixtureBody))
			if w.Code != want {
				t.Fatalf("status=%d want=%d", w.Code, want)
			}
			f.mu.Lock()
			commits := f.commits
			f.deny = false
			f.mu.Unlock()
			if want != 200 && commits != 0 {
				t.Fatal("failed transaction wrote receipt")
			}
			if mode == "denied" {
				w = serveTermination(h.handler, h.request(terminationFixtureBody))
				if w.Code != http.StatusOK {
					t.Fatalf("retry status=%d", w.Code)
				}
			}
		})
	}
}
