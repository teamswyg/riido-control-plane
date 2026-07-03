package main

import (
	"fmt"
	"strings"
)

func renderDoc(p projectionManifest, s sourceManifest) string {
	var b strings.Builder
	b.WriteString("# Figma AI Agent Control-Plane Projection\n\n")
	b.WriteString(generatedNotice + "\n\n")
	b.WriteString("Executable SSOT: [`figma-ai-agent-control-plane-projection.riido.json`](figma-ai-agent-control-plane-projection.riido.json).\n\n")
	fmt.Fprintf(&b, "> Riido task: %s\n\n", p.RiidoTask)
	b.WriteString("This repo does not redefine the Figma top-level UI coverage. It projects the contracts-owned Figma coverage into HTTP/SSE, OpenAPI, and generated-client evidence.\n\n")
	renderSource(&b, p, s)
	renderLimitations(&b, p, s)
	renderAnnotations(&b, s)
	renderInventory(&b, s)
	renderProjectionEntries(&b, p)
	renderAbsorptions(&b, p)
	renderBoundaries(&b, s)
	renderLoop(&b, p.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderSource(b *strings.Builder, p projectionManifest, s sourceManifest) {
	b.WriteString("## Source Contract\n\n")
	fmt.Fprintf(b, "- source repo: `%s`\n", p.Source.Repo)
	fmt.Fprintf(b, "- source manifest: `%s`\n", p.Source.Path)
	fmt.Fprintf(b, "- source id: `%s`\n", p.Source.ID)
	fmt.Fprintf(b, "- Figma file: `%s` `%s`\n\n", s.Figma.FileKey, s.Figma.FileName)
	b.WriteString("The full upstream coverage provenance is mirrored from contracts; limitation-local provenance is preserved separately.\n\n")
	for _, pr := range p.Source.StabilizedBy {
		fmt.Fprintf(b, "- %s\n", pr)
	}
	b.WriteByte('\n')
	renderInspection(b, s.InspectionMethod)
}

func renderInspection(b *strings.Builder, method inspectionMethod) {
	b.WriteString("## Inspection Method\n\n")
	fmt.Fprintf(b, "- id: `%s`\n", method.ID)
	fmt.Fprintf(b, "- page registry: `%s`\n", method.PageRegistryExpression)
	fmt.Fprintf(b, "- top-level child count: `%s`\n", method.TopLevelChildCountExpression)
	b.WriteString("- rule: supporting evidence only; passive pages can be lazy/unloaded, so supporting tools must not redefine page-level child counts.\n")
	if strings.TrimSpace(method.Rule) != "" {
		fmt.Fprintf(b, "- mirrored rule: %s\n", method.Rule)
	}
	b.WriteByte('\n')
}

func renderLoop(b *strings.Builder, loop evidenceLoop) {
	b.WriteString("## Evidence Loop\n\n")
	fmt.Fprintf(b, "- observe: %s\n", loop.Observation)
	fmt.Fprintf(b, "- hypothesis: %s\n", loop.Hypothesis)
	fmt.Fprintf(b, "- execute: %s\n", loop.Execute)
	fmt.Fprintf(b, "- evaluate: %s\n", loop.Evaluate)
	fmt.Fprintf(b, "- retrospective: %s\n", loop.Retrospective)
}
