package riidoaiserver

import "testing"

func TestCoverageTerminationClearsOnlyLocalUnresolvedOnce(t *testing.T) {
	for _, scope := range []string{"local_invocations_quiescent", "unobserved"} {
		t.Run(scope, func(t *testing.T) {
			f, store, id := newCoveredTerminationFixture(t, scope)
			f.loseACK = true
			first, err := store.recordExecutionTermination(t.Context(), id, coverageFixtureTermination())
			if err != nil {
				t.Fatal(err)
			}
			second, err := store.recordExecutionTermination(t.Context(), id, coverageFixtureTermination())
			if err != nil || first != second {
				t.Fatal("lost ACK/retry did not preserve original receipt")
			}
			ledger := coverageFixtureLedger(t, f)
			want := uint64(0)
			if scope == "unobserved" {
				want = 1
			}
			if ledger.Revision != 8 || ledger.Unresolved != want {
				t.Fatalf("receipt changed wrong ledger: revision=%d unresolved=%d", ledger.Revision, ledger.Unresolved)
			}
		})
	}
}

func TestCoverageTerminationRotationKeepsOriginalClaimClosable(t *testing.T) {
	f, store, id := newCoveredTerminationFixture(t, "local_invocations_quiescent")
	f.mu.Lock()
	f.items["DEVICE_POLICY#dcov_fixture|REVISION"]["credential_generation"] = map[string]string{"S": "rotated-credential"}
	f.items["DEVICE_POLICY#dcov_fixture|REVISION"]["revision"] = map[string]string{"N": "8"}
	f.mu.Unlock()
	if _, err := store.recordExecutionTermination(t.Context(), id, coverageFixtureTermination()); err == nil {
		t.Fatal("stale credential authority closed an execution")
	}
	id.coverage.credential = "rotated-credential"
	if _, err := store.recordExecutionTermination(t.Context(), id, coverageFixtureTermination()); err != nil {
		t.Fatal("current credential cannot close its original claim:", err)
	}
	if ledger := coverageFixtureLedger(t, f); ledger.Revision != 9 || ledger.Unresolved != 0 {
		t.Fatal("credential rotation lost the original unresolved execution")
	}
}
