package riidoaiserver

import (
	"strings"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func normalizeRuntimeKind(kind RuntimeKind) RuntimeKind {
	switch RuntimeKind(strings.TrimSpace(string(kind))) {
	case RuntimeKindCodex:
		return RuntimeKindCodex
	case RuntimeKindClaudeCode, RuntimeKind("claude"):
		return RuntimeKindClaudeCode
	case RuntimeKindCursor:
		return RuntimeKindCursor
	case RuntimeKindOpenClaw:
		return RuntimeKindOpenClaw
	default:
		return ""
	}
}

func normalizeRuntimeModels(kind RuntimeKind, models []RuntimeModelRecord) []RuntimeModelRecord {
	out := make([]RuntimeModelRecord, 0, len(models))
	seen := map[string]struct{}{}
	for _, model := range models {
		model.ModelID = strings.TrimSpace(model.ModelID)
		model.Label = strings.TrimSpace(model.Label)
		if model.ModelID == "" {
			continue
		}
		if _, ok := seen[model.ModelID]; ok {
			continue
		}
		seen[model.ModelID] = struct{}{}
		if model.Label == "" {
			model.Label = model.ModelID
		}
		out = append(out, model)
	}
	if len(out) == 0 {
		return []RuntimeModelRecord{defaultRuntimeModel(kind)}
	}
	return out
}

func defaultRuntimeModel(kind RuntimeKind) RuntimeModelRecord {
	switch kind {
	case RuntimeKindCodex:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultCodexModelID, Label: "Codex 기본 모델", IsDefault: true}
	case RuntimeKindClaudeCode:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultClaudeModelID, Label: "Claude Code 기본 모델", IsDefault: true}
	case RuntimeKindCursor:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultCursorModelID, Label: "Cursor Auto", IsDefault: true}
	case RuntimeKindOpenClaw:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultOpenClawModelID, Label: "OpenClaw 기본 모델", IsDefault: true}
	default:
		return RuntimeModelRecord{ModelID: providercatalog.DefaultRuntimeModelID, Label: "기본 모델", IsDefault: true}
	}
}

func runtimeProviderForAIAgentRuntime(kind RuntimeKind) string {
	switch normalizeRuntimeKind(kind) {
	case RuntimeKindCodex:
		return "codex"
	case RuntimeKindClaudeCode:
		return "claude"
	case RuntimeKindCursor:
		return "cursor"
	case RuntimeKindOpenClaw:
		return "openclaw"
	default:
		return ""
	}
}
