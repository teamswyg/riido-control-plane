package main

import "testing"

func TestOperationalReadinessRejectsDaemonWildcardNoTestCommand(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].ID = "daemon_network_disconnect_waiting"
	m.Checks[0].NextCommand = "go test ./cmd/riido -run 'TestDaemon.*(LongPoll|Retry)' -count=1"
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected daemon wildcard next command to fail")
	}
}

func TestOperationalReadinessRejectsPlaceholderNextCommand(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].NextCommand = "RIIDO_E2E_TASK_ID=<task> go run ./tools/localproductacceptance"
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected placeholder next command to fail")
	}
}

func TestOperationalReadinessRejectsDaemonGoTestWithoutRepoScope(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].EvidenceRefs = []evidenceRef{{Kind: "external", Path: "riido-daemon:cmd/riido/test.go"}}
	m.Checks[0].NextCommand = "go test ./cmd/riido -run TestBuildDaemonControlPlaneSaaSUsesDefaultLongPollWait -count=1"
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected daemon go test without repo scope to fail")
	}
}

func TestOperationalReadinessAllowsShellRedirectNextCommand(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].NextCommand = "printf ok > /tmp/riido-readiness.txt"
	if err := verifyChecks("../..", m); err != nil {
		t.Fatalf("expected shell redirect next command to pass: %v", err)
	}
}
