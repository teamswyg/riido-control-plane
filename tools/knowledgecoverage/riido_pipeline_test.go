package main

import "testing"

func TestSourceManifestAcceptsAttestedPrivateRiidoPipeline(t *testing.T) {
	root := t.TempDir()
	source := sourceSSOT{
		Path: "config/policy.riido.json", EvidenceTool: "tools/policy",
		Workflow: "pipelines/check.riido.json", EvidenceArtifact: "policy-evidence",
	}
	mustWrite(t, root, source.Path, riidoSourceManifestFixture())
	mustWrite(t, root, source.Workflow, riidoPipelineFixture("required", "required", "always"))
	if problems := validateSourceManifest(root, source); len(problems) != 0 {
		t.Fatalf("strict riido-ci pipeline rejected: %v", problems)
	}
}

func TestSourceManifestRejectsUnattestedOrUnredactedRiidoPipeline(t *testing.T) {
	for _, test := range []struct{ name, attestation, redaction, runWhen string }{
		{"unattested", "none", "required", "always"},
		{"unredacted", "required", "none", "always"},
		{"not_always", "required", "required", "success"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			source := sourceSSOT{"config/policy.riido.json", "tools/policy", "pipelines/check.riido.json", "policy-evidence"}
			mustWrite(t, root, source.Path, riidoSourceManifestFixture())
			mustWrite(t, root, source.Workflow, riidoPipelineFixture(test.attestation, test.redaction, test.runWhen))
			if problems := validateSourceManifest(root, source); len(problems) == 0 {
				t.Fatal("unsafe riido-ci pipeline must not count as strict evidence")
			}
		})
	}
}

func riidoSourceManifestFixture() string {
	return `{"id":"policy","assertions":["a"],"loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"}}`
}

func riidoPipelineFixture(attestation, redaction, runWhen string) string {
	return `{"schema_version":"riido-ci-pipeline.v1","status":"active","visibility":"private",` +
		`"execution":{"attestation":"` + attestation + `"},"evidence_contract":{"artifact":"policy-evidence"},` +
		`"steps":[{"kind":"shell","command":"go run ./tools/policy -contract config/policy.riido.json -evidence-out out/policy.json"},` +
		`{"kind":"artifact","paths":["out/policy.json"],"redaction":"` + redaction + `","run_when":"` + runWhen + `"}]}`
}
