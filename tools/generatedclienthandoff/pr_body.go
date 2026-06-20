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
