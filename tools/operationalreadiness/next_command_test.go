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
