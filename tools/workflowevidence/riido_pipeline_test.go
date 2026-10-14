package main

import "testing"

func TestEvidenceToolInventoryAcceptsStrictRiidoPipelineRoute(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "tools/report/main.go", "package main\nfunc main(){ _ = \"evidence-out\" }\n")
	mustWrite(t, root, "pipelines/check.riido.json", workflowEvidencePipeline("required", "required", "always"))
	got, err := auditWorkflows(root, manifest{WorkflowRoot: ".github/workflows", PipelineFiles: []string{"pipelines/check.riido.json"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.EvidenceToolCovered != 1 || got.EvidenceToolBound != 1 ||
		len(got.MissingEvidenceTools) != 0 || len(got.MissingEvidenceToolBindings) != 0 {
		t.Fatalf("strict riido-ci route did not bind evidence: %+v", got)
	}
}

func TestEvidenceToolInventoryRejectsUnsafeRiidoPipelineRoute(t *testing.T) {
	for _, test := range []struct{ name, attestation, redaction, runWhen string }{
		{"unattested", "none", "required", "always"},
		{"unredacted", "required", "none", "always"},
		{"not_always", "required", "required", "success"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			mustWrite(t, root, "tools/report/main.go", "package main\nfunc main(){ _ = \"evidence-out\" }\n")
			mustWrite(t, root, "pipelines/check.riido.json", workflowEvidencePipeline(test.attestation, test.redaction, test.runWhen))
			_, err := auditWorkflows(root, manifest{WorkflowRoot: ".github/workflows", PipelineFiles: []string{"pipelines/check.riido.json"}})
			if err == nil {
				t.Fatal("unsafe riido-ci route must fail closed")
			}
		})
	}
}

func workflowEvidencePipeline(attestation, redaction, runWhen string) string {
	return `{"schema_version":"riido-ci-pipeline.v1","status":"active","repo":"riido-control-plane","visibility":"private",` +
		`"execution":{"attestation":"` + attestation + `"},"evidence_contract":{"artifact":"report-evidence"},` +
		`"steps":[{"kind":"shell","command":"go run ./tools/report -evidence-out out/report.json"},` +
		`{"kind":"artifact","paths":["out/report.json"],"redaction":"` + redaction + `","run_when":"` + runWhen + `"}]}`
}
