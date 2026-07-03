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
	if evidence.Source.NotionComment != "39220241-cf7f-81a7-b4ca-001d8d6c5d1d" {
		t.Fatalf("notion comment id = %q", evidence.Source.NotionComment)
	}
	if evidence.Assertions.RawResponseBodiesIncluded || evidence.Assertions.SecretsIncluded {
		t.Fatal("route evidence must not include raw bodies or secrets")
	}
	if !evidence.Assertions.Not404 || !evidence.Assertions.UnauthenticatedProbeOnly {
		t.Fatal("route evidence must prove unauthenticated 401 route existence")
	}
	want := map[string]string{
		"staging":     "already_live_before_this_capture",
		"development": "https://github.com/teamswyg/riido-control-plane/actions/runs/28640542631",
		"production":  "https://github.com/teamswyg/riido-control-plane/actions/runs/28640723634",
	}
	if len(evidence.Deployments) != len(want) {
		t.Fatalf("deployment count = %d", len(evidence.Deployments))
	}
	for _, deployment := range evidence.Deployments {
		run, ok := want[deployment.Environment]
		if !ok {
			t.Fatalf("unexpected environment %q", deployment.Environment)
		}
		if deployment.DeploymentRun != run || deployment.V1Status != 401 || deployment.V2Status != 401 {
			t.Fatalf("bad deployment evidence: %+v", deployment)
		}
	}
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

type deviceDaemonsPluralEvidence struct {
	Redacted bool `json:"redacted"`
	Source   struct {
		NotionComment string `json:"notion_comment"`
	} `json:"source"`
	Deployments []struct {
		Environment   string `json:"environment"`
		DeploymentRun string `json:"deployment_run"`
		V1Status      int    `json:"v1_status"`
		V2Status      int    `json:"v2_status"`
	} `json:"deployments"`
	Assertions struct {
		Not404                    bool `json:"not_404"`
		UnauthenticatedProbeOnly  bool `json:"unauthenticated_probe_only"`
		RawResponseBodiesIncluded bool `json:"raw_response_bodies_included"`
		SecretsIncluded           bool `json:"secrets_included"`
	} `json:"assertions"`
}
