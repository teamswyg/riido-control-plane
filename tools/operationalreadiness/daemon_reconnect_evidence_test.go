package main

import (
	"encoding/json"
	"os"
	"testing"
)

const daemonReconnectEvidence = "docs/30-architecture/evidence/daemon-reconnect-runtimeactor-evidence-2026-07-02.json"

type daemonReconnectFinding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
}

func TestDaemonReconnectEvidencePreservesReleasePartial(t *testing.T) {
	body, err := os.ReadFile("../../" + daemonReconnectEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted bool `json:"redacted"`
		Source   struct {
			Commit string `json:"commit"`
		} `json:"source"`
		Commands []struct {
			Module      string   `json:"module"`
			PackagePath string   `json:"package_path"`
			Status      string   `json:"status"`
			PassedTests []string `json:"passed_tests"`
		} `json:"commands"`
		Findings []daemonReconnectFinding `json:"findings"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || evidence.Source.Commit == "" {
		t.Fatal("daemon reconnect evidence must be redacted and bind daemon commit")
	}
	if len(evidence.Commands) != 3 {
		t.Fatalf("commands = %d, want 3", len(evidence.Commands))
	}
	for _, command := range evidence.Commands {
		if command.Module != "teamswyg/riido-daemon" || command.PackagePath == "" {
			t.Fatalf("command must bind daemon module and package path without private absolute import: %+v", command)
		}
		if command.Status != "passed" || len(command.PassedTests) == 0 {
			t.Fatalf("unexpected command evidence: %+v", command)
		}
	}
	if !hasDaemonReconnectFinding(evidence.Findings, "release_chaos_not_exercised", "partial") {
		t.Fatal("daemon reconnect evidence must keep release chaos partial")
	}
}

func hasDaemonReconnectFinding(findings []daemonReconnectFinding, id, severity string) bool {
	for _, finding := range findings {
		if finding.ID == id && finding.Severity == severity {
			return true
		}
	}
	return false
}
