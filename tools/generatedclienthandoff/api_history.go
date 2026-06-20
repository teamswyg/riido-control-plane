package main

import (
	"fmt"
	"strings"
)

func apiHistory(cfg config, ops []operationRow) string {
	var b strings.Builder
	b.WriteString("// 이 파일은 Riido control-plane API 전달 이력을 기록하는 generated 파일입니다. 직접 수정하지 마세요.\n\n")
	b.WriteString("export interface RiidoControlPlaneAPIHistoryEntry {\n")
	b.WriteString("  readonly sourceCommit: string;\n  readonly sourceRef: string;\n  readonly releasedAt: string;\n")
	b.WriteString("  readonly summary: readonly string[];\n  readonly ssotDecisions: readonly string[];\n")
	b.WriteString("  readonly operations: readonly string[];\n}\n\n")
	b.WriteString("export const riidoControlPlaneAPIHistory = [\n")
	b.WriteString("  {\n")
	fmt.Fprintf(&b, "    sourceCommit: '%s',\n", cfg.SourceCommit)
	fmt.Fprintf(&b, "    sourceRef: '%s',\n", cfg.SourceRef)
	fmt.Fprintf(&b, "    releasedAt: '%s',\n", cfg.GeneratedAt)
	renderHistoryBody(&b, ops)
	b.WriteString("  },\n")
	b.WriteString("] as const satisfies readonly RiidoControlPlaneAPIHistoryEntry[];\n")
	return b.String()
}

func renderHistoryBody(b *strings.Builder, ops []operationRow) {
	b.WriteString("    summary: [\n")
	b.WriteString("      'control-plane OpenAPI/DSL/IR 기준 React Query generated client를 갱신했습니다.',\n")
	b.WriteString("      'PR은 generated handoff이며 riido-client에서 검토 후 merge 여부를 결정합니다.',\n")
	b.WriteString("    ],\n")
	b.WriteString("    ssotDecisions: [\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(b, "      '%s',\n", ts(line))
	}
	b.WriteString("    ],\n")
	b.WriteString("    operations: [\n")
	for _, op := range ops {
		fmt.Fprintf(b, "      '%s',\n", ts(op.GeneratedPath))
	}
	b.WriteString("    ],\n")
}
