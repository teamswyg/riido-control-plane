package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`api-client-delivery.riido.json`](api-client-delivery.riido.json).\n\n")
	b.WriteString("This reader is generated from the generated-client delivery manifest and source evidence.\n\n")
	renderCoverage(&b, result)
	renderSources(&b, m.Sources)
	renderOwners(&b, m.Owners)
	renderDelivery(&b, m)
	renderGenerator(&b, m.Generator)
	renderFigma(&b, m.Figma)
	renderModelCatalog(&b, m.ModelCatalog)
	renderRiskEvidence(&b, result.RiskEvidence)
	renderList(&b, "Lifecycle States", m.Lifecycle)
	renderList(&b, "Non-Goals", m.NonGoals)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderCoverage(b *strings.Builder, r verifyResult) {
	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(b, "Source manifests: `%d`; owners: `%d`; Figma contexts: `%d`; source checks: `%d`; phrase checks: `%d`; forbidden checks: `%d`; risk tests: `%d`.\n\n",
		r.SourceManifests, r.Owners, r.FigmaContexts, r.SourceChecks, r.PhraseChecks, r.ForbiddenChecks, r.RiskTests)
}
