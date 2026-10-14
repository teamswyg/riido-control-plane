package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest) string {
	var b strings.Builder
	b.WriteString(generatedNotice + "\n\n")
	fmt.Fprintf(&b, "# %s\n\n", m.Title)
	b.WriteString("This reader is generated from the health/ready command SSOT.\n\n")
	b.WriteString("## Runtime Contract\n\n")
	fmt.Fprintf(&b, "- Workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- Evidence artifact: `%s`\n", m.EvidenceArtifact)
	fmt.Fprintf(&b, "- Owners: `%s`\n\n", strings.Join(m.OwnerPackages, "`, `"))
	renderEndpoints(&b, m.Endpoints)
	renderLoop(&b, m.Loop)
	return b.String()
}

func renderEndpoints(b *strings.Builder, endpoints []endpointContract) {
	b.WriteString("## Endpoints\n\n| Name | Route | HTTP | Status |\n| --- | --- | ---: | --- |\n")
	for _, endpoint := range endpoints {
		fmt.Fprintf(b, "| `%s` | `%s %s` | `%d` | `%s` |\n", endpoint.Name, endpoint.Method, endpoint.Path, endpoint.HTTPStatus, endpoint.Status)
	}
	b.WriteString("\n")
}
