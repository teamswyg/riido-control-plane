package main

import "testing"

func TestVerifyCaseNamesRejectsDrift(t *testing.T) {
	t.Parallel()
	cases := []caseSpec{{Name: "a"}, {Name: "b"}}
	if err := verifyCaseNames(cases, []caseEvidence{{Name: "a"}}); err == nil {
		t.Fatal("verifyCaseNames accepted count drift")
	}
	if err := verifyCaseNames([]caseSpec{{Name: "a"}}, []caseEvidence{{}}); err == nil {
		t.Fatal("verifyCaseNames accepted empty evidence name")
	}
	if err := verifyCaseNames([]caseSpec{{Name: "b"}}, []caseEvidence{{Name: "a"}}); err == nil {
		t.Fatal("verifyCaseNames accepted missing case")
	}
}

func TestVerifyRejectsCaseNameDrift(t *testing.T) {
	t.Parallel()
	m := manifest{Cases: []caseSpec{{Name: "expected"}}}
	if err := verify(t.TempDir(), m, []caseEvidence{{Name: "other"}}, false); err == nil {
		t.Fatal("verify accepted case name drift")
	}
}

func TestVerifySnapshotResultRejectsDrift(t *testing.T) {
	t.Parallel()
	tc := caseSpec{
		Name: "snapshot", WantHistoryEvents: 3,
		WantAssignmentState: "ready", WantTaskEvents: 3,
	}
	cases := map[string]caseEvidence{
		"history": {Name: "snapshot", HistoryEvents: 2, AssignmentState: "ready", TaskEvents: 3},
		"state":   {Name: "snapshot", HistoryEvents: 3, AssignmentState: "queued", TaskEvents: 3},
		"events":  {Name: "snapshot", HistoryEvents: 3, AssignmentState: "ready", TaskEvents: 2},
	}
	for name, got := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := verifySnapshotResult(tc, got); err == nil {
				t.Fatalf("verifySnapshotResult accepted %s drift", name)
			}
		})
	}
}

func TestVerifySnapshotCaseRejectsExpectationDrift(t *testing.T) {
	t.Parallel()
	tc := testCaseByKind(t, "snapshot")
	tc.Name = "snapshot-expectation-drift"
	tc.WantHistoryEvents = 99
	if _, err := verifySnapshotCase(tc); err == nil {
		t.Fatal("verifySnapshotCase accepted expectation drift")
	}
}

func TestVerifyDocRejectsMissingAndStaleDoc(t *testing.T) {
	t.Parallel()
	m := testManifest(t)
	if err := verifyDoc(t.TempDir(), m); err == nil {
		t.Fatal("verifyDoc accepted missing generated doc")
	}
	stale := m
	stale.GeneratedDoc = "docs/30-architecture/store-snapshot-file-outbox.riido.json"
	if err := verifyDoc(testRepoRoot(t), stale); err == nil {
		t.Fatal("verifyDoc accepted stale generated doc")
	}
}
