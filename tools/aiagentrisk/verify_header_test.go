package main

import (
	"strings"
	"testing"
)

func TestVerifyHeaderRejectsManifestIdentityDrift(t *testing.T) {
	valid := evidenceManifest{
		SchemaVersion: schemaVersion,
		ID:            "control-plane-ai-agent-risk-evidence",
		RiidoTask:     "RIID-4964",
		HumanDoc:      "docs/30-architecture/api-client-delivery.md",
	}
	doc := "linked ai-agent-risk-evidence.riido.json"
	cases := []struct {
		name string
		edit func(*evidenceManifest)
		want string
	}{
		{"schema", func(m *evidenceManifest) { m.SchemaVersion = "v0" }, "schema_version"},
		{"id", func(m *evidenceManifest) { m.ID = "other" }, "unexpected id"},
		{"task", func(m *evidenceManifest) { m.RiidoTask = "RIID-0" }, "riido_task"},
		{"doc", func(m *evidenceManifest) { m.HumanDoc = "README.md" }, "human_doc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := valid
			tc.edit(&item)
			err := verifyHeader(item, doc)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestVerifyHeaderRequiresExecutableManifestLink(t *testing.T) {
	err := verifyHeader(evidenceManifest{
		SchemaVersion: schemaVersion,
		ID:            "control-plane-ai-agent-risk-evidence",
		RiidoTask:     "RIID-4964",
		HumanDoc:      "docs/30-architecture/api-client-delivery.md",
	}, "human-only prose")
	if err == nil || !strings.Contains(err.Error(), "must link") {
		t.Fatalf("expected missing link error, got %v", err)
	}
}
