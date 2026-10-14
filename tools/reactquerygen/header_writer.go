package main

import "strings"

func writeGeneratedHeader(b *strings.Builder) {
	b.WriteString("// 이 파일은 tools/reactquerygen으로 생성됩니다. 직접 수정하지 마세요.\n")
	b.WriteString("// 원본: contracts/ai-agent-client/control-plane-ai-agent-client.openapi.json\n\n")
}
