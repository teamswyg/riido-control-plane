package main

import (
	"fmt"
	"strings"
)

func readme(cfg config, hashes map[string]string, ops []operationRow) string {
	var b strings.Builder
	b.WriteString("# Riido Control Plane React Query Client 전달본\n\n")
	b.WriteString("이 디렉터리는 `teamswyg/riido-control-plane`의 AI Agent client API를 `teamswyg/riido-client`에서 검토할 수 있도록 생성한 전달본입니다.\n\n")
	renderReadmeBasis(&b, cfg, hashes)
	renderReadmeDecisions(&b)
	renderReadmeFiles(&b)
	renderReadmeGeneratedPaths(&b, ops)
	return b.String()
}

func renderReadmeBasis(b *strings.Builder, cfg config, hashes map[string]string) {
	b.WriteString("## 생성 기준\n\n")
	fmt.Fprintf(b, "- source repository: `%s`\n", cfg.SourceRepo)
	fmt.Fprintf(b, "- source ref: `%s`\n", cfg.SourceRef)
	fmt.Fprintf(b, "- source commit: `%s`\n", cfg.SourceCommit)
	fmt.Fprintf(b, "- target branch: `%s`\n", cfg.TargetBranch)
	fmt.Fprintf(b, "- generated at: `%s`\n", cfg.GeneratedAt)
	fmt.Fprintf(b, "- OpenAPI SHA256: `%s`\n\n", hashes["openapi"])
}

func renderReadmeDecisions(b *strings.Builder) {
	b.WriteString("## SSOT 결정사항\n\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(b, "- %s\n", line)
	}
	b.WriteString("\n## 이번 전달본의 변경 기준\n\n")
	b.WriteString("- PR 본문과 이 README는 같은 OpenAPI/DSL/IR 해시에서 생성됩니다.\n")
	b.WriteString("- 변경 파일은 `src/generated/react-query/riido-control-plane/**` 아래 generated artifact로 제한됩니다.\n")
	b.WriteString("- control-plane workflow는 client PR을 열거나 갱신할 수 있지만 자동 merge하지 않습니다.\n\n")
}

func renderReadmeFiles(b *strings.Builder) {
	b.WriteString("## 제공 파일\n\n")
	b.WriteString("| 파일 | 설명 |\n| --- | --- |\n")
	b.WriteString("| `aiAgentClient.ts` | DTO 타입, request 함수, query/mutation helper, core facade |\n")
	b.WriteString("| `aiAgentClient.react.ts` | client component용 React Query hook facade |\n")
	b.WriteString("| `index.ts` | server-safe core barrel export |\n")
	b.WriteString("| `react.ts` | hook wrapper barrel export |\n")
	b.WriteString("| `apiHistory.generated.ts` | 프론트 검토용 변경 이력 |\n")
	b.WriteString("| `contractManifest.generated.ts` | source commit, digest, generated path manifest |\n\n")
}

func renderReadmeGeneratedPaths(b *strings.Builder, ops []operationRow) {
	b.WriteString("## Generated paths\n\n")
	b.WriteString("| Generated path | Operation | HTTP | Lifecycle |\n| --- | --- | --- | --- |\n")
	for _, op := range ops {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s %s` | `%s` |\n", op.GeneratedPath, op.OperationID, op.Method, op.Path, lifecycleLabel(op))
	}
}
