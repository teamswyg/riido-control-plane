package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	fmt.Fprintf(&b, "> Riido task: %s\n\n", m.RiidoTask)
	b.WriteString("Executable SSOT: [`module-decomposition.riido.json`](module-decomposition.riido.json).\n\n")
	b.WriteString("This reader is generated from the package boundary manifest and current Go package surface.\n\n")
	renderCoverage(&b, result)
	renderPackages(&b, m.Packages)
	renderList(&b, "Boundary Rules", m.Rules)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderCoverage(b *strings.Builder, result verifyResult) {
	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(b, "Packages: `%d`; runtime: `%d`; internal: `%d`; tools: `%d`; forbidden import hits: `%d`.\n\n",
		result.PackageCount, result.RuntimePackages, result.InternalPackages,
		result.ToolPackages, result.ForbiddenImportHits)
}
