package main

import "fmt"

func verifyHeader(manifest evidenceManifest, doc string) error {
	if manifest.SchemaVersion != schemaVersion {
		return fmt.Errorf("unexpected schema_version %q", manifest.SchemaVersion)
	}
	if manifest.ID != "control-plane-ai-agent-risk-evidence" {
		return fmt.Errorf("unexpected id %q", manifest.ID)
	}
	if manifest.RiidoTask != "RIID-4964" {
		return fmt.Errorf("unexpected riido_task %q", manifest.RiidoTask)
	}
	if manifest.HumanDoc != "docs/30-architecture/api-client-delivery.md" {
		return fmt.Errorf("unexpected human_doc %q", manifest.HumanDoc)
	}
	if !docMentions(doc, "ai-agent-risk-evidence.riido.json") {
		return fmt.Errorf("human doc must link the executable risk evidence manifest")
	}
	return nil
}
