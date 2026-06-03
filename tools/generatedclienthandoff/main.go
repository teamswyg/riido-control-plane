package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type openAPISpec struct {
	Paths map[string]map[string]operation `json:"paths"`
}

type operation struct {
	OperationID    string         `json:"operationId"`
	Summary        string         `json:"summary"`
	Deprecated     bool           `json:"deprecated"`
	Client         clientMetadata `json:"x-riido-client"`
	Lifecycle      string         `json:"x-riido-lifecycle"`
	Replacement    string         `json:"x-riido-replacement"`
	RemovalHorizon string         `json:"x-riido-removal-horizon"`
}

type clientMetadata struct {
	GeneratedPath string `json:"generated_path"`
}

type operationRow struct {
	Method         string
	Path           string
	OperationID    string
	Summary        string
	GeneratedPath  string
	Deprecated     bool
	Lifecycle      string
	Replacement    string
	RemovalHorizon string
}

type config struct {
	OpenAPI      string
	DSL          string
	IR           string
	Core         string
	React        string
	Out          string
	PRBody       string
	SourceRepo   string
	SourceRef    string
	SourceCommit string
	TargetRepo   string
	TargetBranch string
	GeneratedAt  string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.OpenAPI, "openapi", "", "OpenAPI JSON path")
	flag.StringVar(&cfg.DSL, "dsl", "", "DSL JSON path")
	flag.StringVar(&cfg.IR, "ir", "", "IR JSON path")
	flag.StringVar(&cfg.Core, "core", "", "generated core TypeScript path")
	flag.StringVar(&cfg.React, "react", "", "generated React TypeScript path")
	flag.StringVar(&cfg.Out, "out", "", "output directory")
	flag.StringVar(&cfg.PRBody, "pr-body", "", "optional generated PR body path")
	flag.StringVar(&cfg.SourceRepo, "source-repo", "teamswyg/riido-control-plane", "source repository")
	flag.StringVar(&cfg.SourceRef, "source-ref", "", "source ref or tag")
	flag.StringVar(&cfg.SourceCommit, "source-commit", "", "source commit SHA")
	flag.StringVar(&cfg.TargetRepo, "target-repo", "teamswyg/riido-client", "target repository")
	flag.StringVar(&cfg.TargetBranch, "target-branch", "", "target branch name")
	flag.StringVar(&cfg.GeneratedAt, "generated-at", "", "YYYY-MM-DD generated date")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "generatedclienthandoff:", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	if strings.TrimSpace(cfg.OpenAPI) == "" || strings.TrimSpace(cfg.DSL) == "" || strings.TrimSpace(cfg.IR) == "" ||
		strings.TrimSpace(cfg.Core) == "" || strings.TrimSpace(cfg.React) == "" || strings.TrimSpace(cfg.Out) == "" {
		return errors.New("openapi, dsl, ir, core, react, and out are required")
	}
	if strings.TrimSpace(cfg.SourceCommit) == "" {
		return errors.New("source-commit is required")
	}
	if strings.TrimSpace(cfg.TargetBranch) == "" {
		return errors.New("target-branch is required")
	}
	if strings.TrimSpace(cfg.SourceRef) == "" {
		cfg.SourceRef = cfg.SourceCommit
	}
	if strings.TrimSpace(cfg.GeneratedAt) == "" {
		cfg.GeneratedAt = time.Now().UTC().Format("2006-01-02")
	}
	ops, err := readOperations(cfg.OpenAPI)
	if err != nil {
		return err
	}
	hashes, err := fileHashes(map[string]string{
		"openapi": cfg.OpenAPI,
		"dsl":     cfg.DSL,
		"ir":      cfg.IR,
		"core":    cfg.Core,
		"react":   cfg.React,
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.Out, 0o755); err != nil {
		return err
	}
	files := map[string]string{
		"README.generated.md":           readme(cfg, hashes, ops),
		"apiHistory.generated.ts":       apiHistory(cfg, ops),
		"contractManifest.generated.ts": contractManifest(cfg, hashes, ops),
		"index.ts":                      "export * from './aiAgentClient';\nexport * from './apiHistory.generated';\nexport * from './contractManifest.generated';\n",
		"react.ts":                      "export * from './aiAgentClient.react';\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(cfg.Out, name), []byte(body), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	if strings.TrimSpace(cfg.PRBody) != "" {
		if err := os.WriteFile(cfg.PRBody, []byte(prBody(cfg, hashes, ops)), 0o644); err != nil {
			return fmt.Errorf("write pr body: %w", err)
		}
	}
	return nil
}

func readOperations(path string) ([]operationRow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec openAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	var rows []operationRow
	for path, methods := range spec.Paths {
		for method, op := range methods {
			rows = append(rows, operationRow{
				Method:         strings.ToUpper(method),
				Path:           path,
				OperationID:    op.OperationID,
				Summary:        op.Summary,
				GeneratedPath:  op.Client.GeneratedPath,
				Deprecated:     op.Deprecated,
				Lifecycle:      op.Lifecycle,
				Replacement:    op.Replacement,
				RemovalHorizon: op.RemovalHorizon,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].GeneratedPath == rows[j].GeneratedPath {
			return rows[i].OperationID < rows[j].OperationID
		}
		return rows[i].GeneratedPath < rows[j].GeneratedPath
	})
	return rows, nil
}

func fileHashes(paths map[string]string) (map[string]string, error) {
	out := map[string]string{}
	for name, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		out[name] = hex.EncodeToString(sum[:])
	}
	return out, nil
}

func readme(cfg config, hashes map[string]string, ops []operationRow) string {
	var b strings.Builder
	b.WriteString("# Riido Control Plane React Query Client 전달본\n\n")
	b.WriteString("이 디렉터리는 `teamswyg/riido-control-plane`의 AI Agent client API를 `teamswyg/riido-client`에서 검토할 수 있도록 생성한 전달본입니다.\n\n")
	b.WriteString("## 생성 기준\n\n")
	fmt.Fprintf(&b, "- source repository: `%s`\n", cfg.SourceRepo)
	fmt.Fprintf(&b, "- source ref: `%s`\n", cfg.SourceRef)
	fmt.Fprintf(&b, "- source commit: `%s`\n", cfg.SourceCommit)
	fmt.Fprintf(&b, "- target branch: `%s`\n", cfg.TargetBranch)
	fmt.Fprintf(&b, "- generated at: `%s`\n", cfg.GeneratedAt)
	fmt.Fprintf(&b, "- OpenAPI SHA256: `%s`\n\n", hashes["openapi"])
	b.WriteString("## SSOT 결정사항\n\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(&b, "- %s\n", line)
	}
	b.WriteString("\n## 이번 전달본의 변경 기준\n\n")
	b.WriteString("- PR 본문과 이 README는 같은 OpenAPI/DSL/IR 해시에서 생성됩니다.\n")
	b.WriteString("- 변경 파일은 `src/generated/react-query/riido-control-plane/**` 아래 generated artifact로 제한됩니다.\n")
	b.WriteString("- control-plane workflow는 client PR을 열거나 갱신할 수 있지만 자동 merge하지 않습니다.\n\n")
	b.WriteString("## 제공 파일\n\n")
	b.WriteString("| 파일 | 설명 |\n| --- | --- |\n")
	b.WriteString("| `aiAgentClient.ts` | DTO 타입, request 함수, query/mutation helper, core facade |\n")
	b.WriteString("| `aiAgentClient.react.ts` | client component용 React Query hook facade |\n")
	b.WriteString("| `index.ts` | server-safe core barrel export |\n")
	b.WriteString("| `react.ts` | hook wrapper barrel export |\n")
	b.WriteString("| `apiHistory.generated.ts` | 프론트 검토용 변경 이력 |\n")
	b.WriteString("| `contractManifest.generated.ts` | source commit, digest, generated path manifest |\n\n")
	b.WriteString("## Generated paths\n\n")
	b.WriteString("| Generated path | Operation | HTTP | Lifecycle |\n| --- | --- | --- | --- |\n")
	for _, op := range ops {
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s %s` | `%s` |\n", op.GeneratedPath, op.OperationID, op.Method, op.Path, lifecycleLabel(op))
	}
	return b.String()
}

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
	b.WriteString("    summary: [\n")
	b.WriteString("      'control-plane OpenAPI/DSL/IR 기준 React Query generated client를 갱신했습니다.',\n")
	b.WriteString("      'PR은 generated handoff이며 riido-client에서 검토 후 merge 여부를 결정합니다.',\n")
	b.WriteString("    ],\n")
	b.WriteString("    ssotDecisions: [\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(&b, "      '%s',\n", ts(line))
	}
	b.WriteString("    ],\n")
	b.WriteString("    operations: [\n")
	for _, op := range ops {
		fmt.Fprintf(&b, "      '%s',\n", ts(op.GeneratedPath))
	}
	b.WriteString("    ],\n")
	b.WriteString("  },\n")
	b.WriteString("] as const satisfies readonly RiidoControlPlaneAPIHistoryEntry[];\n")
	return b.String()
}

func contractManifest(cfg config, hashes map[string]string, ops []operationRow) string {
	var b strings.Builder
	b.WriteString("// 이 파일은 Riido control-plane API 전달 manifest입니다. 직접 수정하지 마세요.\n\n")
	b.WriteString("export const riidoControlPlaneContractManifest = {\n")
	b.WriteString("  schemaVersion: 'riido-control-plane-client-delivery-manifest.v2',\n")
	fmt.Fprintf(&b, "  targetRepository: '%s',\n", cfg.TargetRepo)
	fmt.Fprintf(&b, "  targetBranch: '%s',\n", cfg.TargetBranch)
	b.WriteString("  outputPath: 'src/generated/react-query/riido-control-plane',\n")
	fmt.Fprintf(&b, "  sourceRepository: '%s',\n", cfg.SourceRepo)
	fmt.Fprintf(&b, "  sourceRef: '%s',\n", cfg.SourceRef)
	fmt.Fprintf(&b, "  sourceCommit: '%s',\n", cfg.SourceCommit)
	b.WriteString("  sourceOpenAPIPath: 'contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json',\n")
	fmt.Fprintf(&b, "  sourceOpenAPISHA256: '%s',\n", hashes["openapi"])
	fmt.Fprintf(&b, "  sourceDSLSHA256: '%s',\n", hashes["dsl"])
	fmt.Fprintf(&b, "  sourceIRSHA256: '%s',\n", hashes["ir"])
	b.WriteString("  generator: {\n")
	b.WriteString("    name: 'tools/reactquerygen + tools/generatedclienthandoff',\n")
	b.WriteString("    owner: 'teamswyg/riido-control-plane',\n")
	b.WriteString("  },\n")
	b.WriteString("  generatedOutputs: {\n")
	b.WriteString("    coreEntry: 'aiAgentClient.ts',\n")
	b.WriteString("    reactEntry: 'aiAgentClient.react.ts',\n")
	b.WriteString("    indexEntry: 'index.ts',\n")
	b.WriteString("    reactBarrelEntry: 'react.ts',\n")
	fmt.Fprintf(&b, "    coreSHA256: '%s',\n", hashes["core"])
	fmt.Fprintf(&b, "    reactSHA256: '%s',\n", hashes["react"])
	b.WriteString("  },\n")
	fmt.Fprintf(&b, "  generatedAt: '%s',\n", cfg.GeneratedAt)
	b.WriteString("  ssotDecisions: [\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(&b, "    '%s',\n", ts(line))
	}
	b.WriteString("  ],\n")
	b.WriteString("  operations: [\n")
	for _, op := range ops {
		fmt.Fprintf(&b, "    { generatedPath: '%s', operationId: '%s', method: '%s', path: '%s', deprecated: %t%s },\n", ts(op.GeneratedPath), ts(op.OperationID), op.Method, ts(op.Path), op.Deprecated, operationLifecycleFields(op))
	}
	b.WriteString("  ],\n")
	b.WriteString("} as const;\n")
	return b.String()
}

func prBody(cfg config, hashes map[string]string, ops []operationRow) string {
	var b strings.Builder
	b.WriteString("## 변경사항\n\n")
	fmt.Fprintf(&b, "- `%s` `%s` 기준으로 control-plane React Query generated client를 갱신했습니다.\n", cfg.SourceRepo, cfg.SourceCommit)
	fmt.Fprintf(&b, "- OpenAPI SHA256: `%s`\n", hashes["openapi"])
	fmt.Fprintf(&b, "- generated operation 수: `%d`\n", len(ops))
	b.WriteString("- 변경 대상은 `src/generated/react-query/riido-control-plane/**` generated artifact입니다.\n\n")
	b.WriteString("## SSOT 기반 결정사항\n\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(&b, "- %s\n", line)
	}
	b.WriteString("\n## Generated paths\n\n")
	for _, op := range ops {
		fmt.Fprintf(&b, "- `%s` -> `%s %s` (`%s`, lifecycle: `%s`)\n", op.GeneratedPath, op.Method, op.Path, op.OperationID, lifecycleLabel(op))
	}
	if notes := lifecycleNotes(ops); len(notes) > 0 {
		b.WriteString("\n## Lifecycle / Deprecation\n\n")
		for _, note := range notes {
			fmt.Fprintf(&b, "- %s\n", note)
		}
	}
	b.WriteString("\n## 검증\n\n")
	b.WriteString("- `go test ./tools/reactquerygen -count=1`\n")
	b.WriteString("- `go run ./tools/reactquerygen ...`\n")
	b.WriteString("- `go run ./tools/generatedclienthandoff ...`\n")
	b.WriteString("- `pnpm exec prettier --check src/generated/react-query/riido-control-plane`\n")
	b.WriteString("- `pnpm run type-check`\n")
	b.WriteString("- workflow guard: generated path allowlist only\n\n")
	b.WriteString("## 참고\n\n")
	b.WriteString("이 PR은 generated handoff입니다. control-plane workflow가 PR을 열거나 갱신하지만 자동 merge하지 않습니다.\n")
	return b.String()
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

func lifecycleLabel(op operationRow) string {
	if strings.TrimSpace(op.Lifecycle) != "" {
		return op.Lifecycle
	}
	if op.Deprecated {
		return "deprecated"
	}
	return "not_declared"
}

func operationLifecycleFields(op operationRow) string {
	var parts []string
	if strings.TrimSpace(op.Lifecycle) != "" {
		parts = append(parts, fmt.Sprintf("lifecycle: '%s'", ts(op.Lifecycle)))
	}
	if strings.TrimSpace(op.Replacement) != "" {
		parts = append(parts, fmt.Sprintf("replacement: '%s'", ts(op.Replacement)))
	}
	if strings.TrimSpace(op.RemovalHorizon) != "" {
		parts = append(parts, fmt.Sprintf("removalHorizon: '%s'", ts(op.RemovalHorizon)))
	}
	if len(parts) == 0 {
		return ""
	}
	return ", " + strings.Join(parts, ", ")
}

func lifecycleNotes(ops []operationRow) []string {
	var notes []string
	for _, op := range ops {
		if !op.Deprecated && strings.TrimSpace(op.Lifecycle) == "" &&
			strings.TrimSpace(op.Replacement) == "" && strings.TrimSpace(op.RemovalHorizon) == "" {
			continue
		}
		var attrs []string
		if op.Deprecated {
			attrs = append(attrs, "deprecated")
		}
		if strings.TrimSpace(op.Lifecycle) != "" {
			attrs = append(attrs, "lifecycle="+op.Lifecycle)
		}
		if strings.TrimSpace(op.Replacement) != "" {
			attrs = append(attrs, "replacement="+op.Replacement)
		}
		if strings.TrimSpace(op.RemovalHorizon) != "" {
			attrs = append(attrs, "removal_horizon="+op.RemovalHorizon)
		}
		notes = append(notes, fmt.Sprintf("`%s`: %s", op.GeneratedPath, strings.Join(attrs, ", ")))
	}
	return notes
}

func ts(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}
