package main

import (
	"fmt"
	"strings"
)

func renderDoc(m manifest, result verifyResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", m.Title, generatedNotice)
	b.WriteString("Executable SSOT: [`pre-commit-baseline.riido.json`](pre-commit-baseline.riido.json).\n\n")
	fmt.Fprintf(&b, "- pre-commit config: `%s`\n", m.PreCommitConfig)
	fmt.Fprintf(&b, "- workflow: `%s`\n", m.Workflow)
	fmt.Fprintf(&b, "- evidence artifact: `%s`\n", m.Evidence)
	fmt.Fprintf(&b, "- evidence ttl hours: `%d`\n", m.EvidenceTTL)
	fmt.Fprintf(&b, "- workflow scheduled: `%t`\n", result.WorkflowScheduled)
	fmt.Fprintf(&b, "- hooks: `%d`\n", result.Hooks)
	fmt.Fprintf(&b, "- scripts: `%d`\n", result.Scripts)
	fmt.Fprintf(&b, "- phrase checks: `%d`\n\n", result.PhraseChecks)
	renderHooks(&b, m.Hooks)
	renderScripts(&b, m.Scripts)
	renderLoop(&b, m.Loop)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderHooks(b *strings.Builder, hooks []checkBlock) {
	b.WriteString("## Hooks\n\n| Hook | Summary | Phrases |\n| --- | --- | ---: |\n")
	for _, hook := range hooks {
		fmt.Fprintf(b, "| `%s` | %s | `%d` |\n", hook.ID, hook.Summary, len(hook.Contains))
	}
	b.WriteString("\n")
}
