package main

import (
	"fmt"
	"strings"
)

func contractManifest(cfg config, hashes map[string]string, ops []operationRow) string {
	var b strings.Builder
	b.WriteString("// 이 파일은 Riido control-plane API 전달 manifest입니다. 직접 수정하지 마세요.\n\n")
	b.WriteString("export const riidoControlPlaneContractManifest = {\n")
	b.WriteString("  schemaVersion: 'riido-control-plane-client-delivery-manifest.v2',\n")
	renderManifestSource(&b, cfg, hashes)
	renderManifestOutputs(&b, cfg, hashes)
	renderManifestOperations(&b, ops)
	b.WriteString("} as const;\n")
	return b.String()
}

func renderManifestSource(b *strings.Builder, cfg config, hashes map[string]string) {
	fmt.Fprintf(b, "  targetRepository: '%s',\n", cfg.TargetRepo)
	fmt.Fprintf(b, "  targetBranch: '%s',\n", cfg.TargetBranch)
	b.WriteString("  outputPath: 'src/generated/react-query/riido-control-plane',\n")
	fmt.Fprintf(b, "  sourceRepository: '%s',\n", cfg.SourceRepo)
	fmt.Fprintf(b, "  sourceRef: '%s',\n", cfg.SourceRef)
	fmt.Fprintf(b, "  sourceCommit: '%s',\n", cfg.SourceCommit)
	b.WriteString("  sourceOpenAPIPath: 'contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json',\n")
	fmt.Fprintf(b, "  sourceOpenAPISHA256: '%s',\n", hashes["openapi"])
	fmt.Fprintf(b, "  sourceDSLSHA256: '%s',\n", hashes["dsl"])
	fmt.Fprintf(b, "  sourceIRSHA256: '%s',\n", hashes["ir"])
}

func renderManifestOutputs(b *strings.Builder, cfg config, hashes map[string]string) {
	b.WriteString("  generator: {\n")
	b.WriteString("    name: 'tools/reactquerygen + tools/generatedclienthandoff',\n")
	b.WriteString("    owner: 'teamswyg/riido-control-plane',\n")
	b.WriteString("  },\n")
	b.WriteString("  generatedOutputs: {\n")
	b.WriteString("    coreEntry: 'aiAgentClient.ts',\n")
	b.WriteString("    reactEntry: 'aiAgentClient.react.ts',\n")
	b.WriteString("    indexEntry: 'index.ts',\n")
	b.WriteString("    reactBarrelEntry: 'react.ts',\n")
	fmt.Fprintf(b, "    coreSHA256: '%s',\n", hashes["core"])
	fmt.Fprintf(b, "    reactSHA256: '%s',\n", hashes["react"])
	b.WriteString("  },\n")
	fmt.Fprintf(b, "  generatedAt: '%s',\n", cfg.GeneratedAt)
	b.WriteString("  ssotDecisions: [\n")
	for _, line := range ssotDecisionLines() {
		fmt.Fprintf(b, "    '%s',\n", ts(line))
	}
	b.WriteString("  ],\n")
}

func renderManifestOperations(b *strings.Builder, ops []operationRow) {
	b.WriteString("  operations: [\n")
	for _, op := range ops {
		fmt.Fprintf(b, "    { generatedPath: '%s', operationId: '%s', method: '%s', path: '%s', deprecated: %t%s },\n", ts(op.GeneratedPath), ts(op.OperationID), op.Method, ts(op.Path), op.Deprecated, operationLifecycleFields(op))
	}
	b.WriteString("  ],\n")
}
