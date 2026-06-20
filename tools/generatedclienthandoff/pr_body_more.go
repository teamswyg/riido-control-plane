package main

import (
	"fmt"
	"strings"
)

func renderChangeSections(b *strings.Builder, sections []changeSection) {
	for _, section := range sections {
		fmt.Fprintf(b, "### %s\n\n", section.Title)
		for _, line := range section.Lines {
			fmt.Fprintf(b, "- %s\n", line)
		}
		b.WriteString("\n")
	}
}

func renderPRBodyGeneratedPaths(b *strings.Builder, ops []operationRow) {
	b.WriteString("\n## Generated paths\n\n")
	for _, op := range ops {
		fmt.Fprintf(b, "- `%s` -> `%s %s` (`%s`, lifecycle: `%s`)\n", op.GeneratedPath, op.Method, op.Path, op.OperationID, lifecycleLabel(op))
	}
}

func renderPRBodyLifecycle(b *strings.Builder, ops []operationRow) {
	if notes := lifecycleNotes(ops); len(notes) > 0 {
		b.WriteString("\n## Lifecycle / Deprecation\n\n")
		for _, note := range notes {
			fmt.Fprintf(b, "- %s\n", note)
		}
	}
}

func renderPRBodyVerification(b *strings.Builder) {
	b.WriteString("\n## 검증\n\n")
	b.WriteString("- `go test ./tools/reactquerygen -count=1`\n")
	b.WriteString("- `go run ./tools/reactquerygen ...`\n")
	b.WriteString("- `go run ./tools/generatedclienthandoff ...`\n")
	b.WriteString("- `pnpm exec prettier --check src/generated/react-query/riido-control-plane`\n")
	b.WriteString("- `pnpm run type-check`\n")
	b.WriteString("- workflow guard: generated path allowlist only\n\n")
	b.WriteString("## 참고\n\n")
	b.WriteString("이 PR은 generated handoff입니다. control-plane workflow가 PR을 열거나 갱신하지만 자동 merge하지 않습니다.\n")
}
