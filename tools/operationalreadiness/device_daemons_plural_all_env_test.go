package main

import (
	"encoding/json"
	"os"
	"testing"
)

const deviceDaemonsPluralAllEnvEvidence = "docs/30-architecture/evidence/device-daemons-plural-all-env-route-2026-07-03.json"

func TestDeviceDaemonsPluralAllEnvEvidence(t *testing.T) {
	evidence := loadDeviceDaemonsPluralEvidence(t)
	if !evidence.Redacted {
		t.Fatal("plural daemon route evidence must be redacted")
	}
	if evidence.Source.NotionComment != "39220241-cf7f-8129-b12d-001d5b689249" {
		t.Fatalf("notion comment id = %q", evidence.Source.NotionComment)
	}
	if !deviceDaemonsPluralHasRelatedComment(evidence, "39220241-cf7f-8171-9aea-001dcb582334") {
		t.Fatal("latest Notion reconfirmation comment is not linked")
	}
	if evidence.Assertions.RawResponseBodiesIncluded || evidence.Assertions.SecretsIncluded {
		t.Fatal("route evidence must not include raw bodies or secrets")
	}
	if !evidence.Assertions.Not404 || !evidence.Assertions.UnauthenticatedProbeOnly {
		t.Fatal("route evidence must prove unauthenticated 401 route existence")
	}
	want := map[string]string{
		"staging":     "https://github.com/teamswyg/riido-control-plane/actions/runs/28643283686",
		"development": "https://github.com/teamswyg/riido-control-plane/actions/runs/28643517509",
		"production":  "https://github.com/teamswyg/riido-control-plane/actions/runs/28643749629",
	}
	if len(evidence.Deployments) != len(want) {
		t.Fatalf("deployment count = %d", len(evidence.Deployments))
	}
	for _, deployment := range evidence.Deployments {
		run, ok := want[deployment.Environment]
		if !ok {
			t.Fatalf("unexpected environment %q", deployment.Environment)
		}
		if deployment.DeploymentRun != run ||
			deployment.HealthzStatus != 200 ||
			deployment.ReadyzStatus != 200 ||
			deployment.V1Status != 401 ||
			deployment.V2Status != 401 {
			t.Fatalf("bad deployment evidence: %+v", deployment)
		}
	}
}

func deviceDaemonsPluralHasRelatedComment(evidence deviceDaemonsPluralEvidence, id string) bool {
	for _, related := range evidence.Source.RelatedNotionComments {
		if related == id {
			return true
		}
	}
	return false
}

func loadDeviceDaemonsPluralEvidence(t *testing.T) deviceDaemonsPluralEvidence {
	t.Helper()
	body, err := os.ReadFile("../../" + deviceDaemonsPluralAllEnvEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence deviceDaemonsPluralEvidence
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	return evidence
}
