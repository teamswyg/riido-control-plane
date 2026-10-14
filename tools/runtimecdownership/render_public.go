package main

import (
	"fmt"
	"strings"
)

func renderPublicBoundary(b *strings.Builder, m manifest, result verifyResult) {
	b.WriteString("## Public Boundary\n")
	renderPublicExport(b, m)
	renderSurfaceScan(b, m)
	renderConfigPolicy(b, m, result)
	renderSensitiveGuard(b, m)
	renderOperationalPolicy(b, m)
	renderActivationGate(b, m)
}

func renderPublicExport(b *strings.Builder, m manifest) {
	b.WriteString("### Public Export Contract\n")
	fmt.Fprintf(b, "RIID-4835 keeps `%s` as canonical owner and `%s` as infra-awareness owner.\n\n",
		m.PublicExport.CanonicalOwner, m.PublicExport.InfraAwarenessOwner)
	fmt.Fprintf(b, "Allowed public export categories: `%d`; forbidden categories: `%d`.\n\n",
		len(m.PublicExport.AllowedPublicExports), len(m.PublicExport.ForbiddenPublicExports))
	b.WriteString("Image values are deliberately not in that public export set.\n\n")
}

func renderSurfaceScan(b *strings.Builder, m manifest) {
	b.WriteString("### Public Surface Scan\n")
	fmt.Fprintf(b, "RIID-4836 scans `%d` public paths and `%d` forbidden pattern classes.\n\n",
		len(m.PublicSurfaceScan.ScopePaths), len(m.PublicSurfaceScan.ForbiddenRegexes)+len(m.PublicSurfaceScan.ForbiddenLiterals))
	b.WriteString("The generated reader describes categories without publishing live host or account values.\n\n")
}

func renderConfigPolicy(b *strings.Builder, m manifest, result verifyResult) {
	b.WriteString("### Public Config Key Minimization\n")
	fmt.Fprintf(b, "RIID-4839 allows `%d` stable deploy/smoke key names only in the canonical manifest and workflows.\n\n",
		result.StableKeyCount)
	b.WriteString("Reader docs link to the manifest instead of repeating exact CD key lists.\n\n")
}

func renderSensitiveGuard(b *strings.Builder, m manifest) {
	b.WriteString("### Public Sensitive Surface Guard\n")
	fmt.Fprintf(b, "RIID-4842 treats public key names as a sensitivity budget: `%t`.\n\n",
		m.PublicSensitiveSurfaceGuard.PublicKeyNamesAreSensitive)
	b.WriteString("Broad summary docs must not become canonical deploy or smoke key-list sources.\n\n")
}

func renderOperationalPolicy(b *strings.Builder, m manifest) {
	b.WriteString("### Public Operational Detail Minimization\n")
	fmt.Fprintf(b, "RIID-4853 rule: %s\n\n", m.PublicOperationalDetailMinimization.Rule)
}

func renderActivationGate(b *strings.Builder, m manifest) {
	b.WriteString("### CodeDeploy Activation Gate\n")
	status := strings.ReplaceAll(m.CodeDeployActivationGate.Status, "-", " ")
	fmt.Fprintf(b, "RIID-4855 keeps CodeDeploy blue/green operator/environment gated (`%s`).\n\n", status)
	b.WriteString("CodeDeploy activation is not an infra-owned deployment action.\n\n")
}
