package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/rendertext"
)

func renderDeviceClientGuidance(b *strings.Builder, guidance deviceClientGuidance) {
	b.WriteString("\n## Device Client Guidance\n\n")
	fmt.Fprintf(b, "- read endpoint: `%s`\n", guidance.ReadEndpoint)
	fmt.Fprintf(b, "- purpose: %s\n", guidance.Purpose)
	rendertext.List(b, "Device Guidance Rules", guidance.Rules)
}
