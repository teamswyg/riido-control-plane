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

func ssotDecisionLines() []string {
	return []string{
		"AI Agent generated client는 control-plane OpenAPI의 `x-riido-client` metadata에서 facade/query/mutation 경로를 생성합니다.",
		"`riido.v2.aiAgent.*`는 workspace-scoped 새 표면이고, v1은 UI 테스트 호환 표면으로 유지됩니다.",
		"브라우저/desktop webview client는 `aiAgentToken`을 `X-Riido-AI-Agent-Token`으로 전달합니다.",
		"`team_id`, `teamId`, Open API key는 AI Agent generated client request와 daemon DevicePrincipal 인증/인가 입력이 아닙니다.",
		"task 배정 request는 `agent_id`를 보내고, task title/body/branch/repository prompt 조립은 control-plane server-to-server 흐름이 담당합니다.",
		"task thread 화면은 먼저 cold collection을 조회하고, `active_stream`이 있을 때만 SSE stream을 연결합니다.",
		"온보딩의 리도/영실/홍도/지원은 template entity가 아니라 agent 생성을 돕는 fixture입니다.",
		"`assignment_ready`는 daemon/runtime handoff 상태이며 busy-agent queue 표시가 아닙니다.",
	}
}

func ts(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

func tsUnescape(s string) string {
	return strings.ReplaceAll(s, "\\'", "'")
}
