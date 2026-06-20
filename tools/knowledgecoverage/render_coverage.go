package main

import (
	"fmt"
	"strings"
)

func renderCoverageTable(b *strings.Builder, e evidence) {
	b.WriteString("## Coverage\n\n| Class | Count |\n| --- | ---: |\n")
	fmt.Fprintf(b, "| Generated reader docs | %d |\n", e.GeneratedCount)
	fmt.Fprintf(b, "| Generated reader docs with tool | %d |\n", e.GeneratedToolCount)
	fmt.Fprintf(b, "| Generated reader docs with CI evidence | %d |\n", e.GeneratedEvidenceWorkflowCount)
	fmt.Fprintf(b, "| Generated reader docs with declared workflow evidence | %d |\n",
		e.GeneratedDeclaredWorkflowCount)
	fmt.Fprintf(b, "| Generated reader docs with manifest evidence_tool | %d |\n",
		e.GeneratedManifestEvidenceTool)
	fmt.Fprintf(b, "| Generated reader docs with uploaded evidence artifact | %d |\n",
		e.GeneratedArtifactBindingCount)
	fmt.Fprintf(b, "| Direct SSOT docs | %d |\n", e.DirectSSOTCount)
	fmt.Fprintf(b, "| Direct SSOT docs with evidence loop | %d |\n", e.DirectLoopCount)
	fmt.Fprintf(b, "| Direct SSOT docs with CI evidence | %d |\n", e.DirectEvidenceWorkflowCount)
	fmt.Fprintf(b, "| Standalone executable manifests | %d |\n", e.StandaloneManifestCount)
	fmt.Fprintf(b, "| Standalone executable manifests with strict CI evidence | %d |\n",
		e.StandaloneManifestBindingCount)
	fmt.Fprintf(b, "| Source SSOT manifests | %d |\n", e.SourceManifestCount)
	fmt.Fprintf(b, "| Source SSOT manifests with strict CI evidence | %d |\n",
		e.SourceManifestBindingCount)
	fmt.Fprintf(b, "| Contract artifacts | %d |\n", e.ContractArtifactCount)
	fmt.Fprintf(b, "| Contract artifacts with owner evidence | %d |\n",
		e.ContractArtifactBindingCount)
	fmt.Fprintf(b, "| Registered manual docs | %d |\n", e.ManualCount)
	fmt.Fprintf(b, "| Scanned docs | %d |\n\n", e.ScannedCount)
}
