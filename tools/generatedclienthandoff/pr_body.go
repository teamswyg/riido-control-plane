package main

import (
	"fmt"
	"strings"
)

func prBody(cfg config, hashes map[string]string, ops []operationRow, previous previousManifest) string {
	var b strings.Builder
	b.WriteString("## 변경사항\n\n")
	fmt.Fprintf(&b, "- `%s` `%s` 기준으로 control-plane React Query generated client를 갱신했습니다.\n", cfg.SourceRepo, cfg.SourceCommit)
	fmt.Fprintf(&b, "- OpenAPI SHA256: `%s`\n", hashes["openapi"])
	fmt.Fprintf(&b, "- generated operation 수: `%d`\n", len(ops))
	b.WriteString("- 변경 대상은 `src/generated/react-query/riido-control-plane/**` generated artifact입니다.\n\n")
	renderPRBodySummary(&b, previous, ops)
	renderPRBodyDecisions(&b)
	renderPRBodyGeneratedPaths(&b, ops)
	renderPRBodyLifecycle(&b, ops)
	renderPRBodyVerification(&b)
	return b.String()
}

func renderPRBodySummary(b *strings.Builder, previous previousManifest, ops []operationRow) {
	b.WriteString("## 변경 요약\n\n")
	for _, line := range changeSummaryLines(previous, ops) {
		fmt.Fprintf(b, "- %s\n", line)
	}
	if detail := changeSummaryDetails(previous, ops); len(detail) > 0 {
		b.WriteString("\n")
		renderChangeSections(b, detail)
	} else {
		b.WriteString("\n")
	}
}

func renderPRBodyDecisions(b *strings.Builder) {
	b.WriteString("## SSOT 기반 결정사항\n\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(b, "- %s\n", line)
	}
}

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
	for _, line := range []string{"`go test ./tools/reactquerygen -count=1`", "`go run ./tools/reactquerygen ...`", "`go run ./tools/generatedclienthandoff ...`", "`pnpm exec prettier --check src/generated/react-query/riido-control-plane`", "`pnpm run type-check`", "workflow guard: generated path allowlist only"} {
		fmt.Fprintf(b, "- %s\n", line)
	}
	b.WriteString("\n## 참고\n\n이 PR은 generated handoff입니다. control-plane workflow가 PR을 열거나 갱신하지만 자동 merge하지 않습니다.\n")
}
