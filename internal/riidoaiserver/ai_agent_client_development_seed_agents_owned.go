package riidoaiserver

import (
	"time"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func developmentOwnedCodexAgent(now time.Time) AgentClientRecord {
	return AgentClientRecord{
		AgentID:             "agent-owned-codex",
		OwnerPrincipalID:    "user-1",
		WorkspaceID:         defaultAIAgentClientWorkspaceID,
		Name:                "Codex 리뷰어",
		ProfileThumbnailURL: "https://cdn.riido.io/dev/ai-agents/codex-reviewer.png",
		Description:         "코드 변경 위험을 먼저 보는 리뷰 에이전트",
		Instruction:         "코드 변경의 위험과 검증 근거를 우선 확인합니다.",
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-codex-dev",
		RuntimeKind:         RuntimeKindCodex,
		ModelID:             providercatalog.DefaultCodexModelID,
		ModelLabel:          "Codex 기본 모델",
		WorkStatus:          AgentWorkStatusRunning,
		Editability:         AgentEditabilityBlockedAssignedTasks,
		AssignedTaskCount:   1,
		CreatedAt:           now.Add(-72 * time.Hour),
		UpdatedAt:           now.Add(-6 * time.Hour),
	}
}

func developmentOwnedClaudeAgent(now time.Time) AgentClientRecord {
	return AgentClientRecord{
		AgentID:             "agent-owned-claude",
		OwnerPrincipalID:    "user-1",
		WorkspaceID:         defaultAIAgentClientWorkspaceID,
		Name:                "Claude 설계 보조",
		ProfileThumbnailURL: "https://cdn.riido.io/dev/ai-agents/claude-designer.png",
		Description:         "기획 의도를 구현 범위로 정리하는 설계 에이전트",
		Instruction:         "기획 의도와 도메인 정책을 먼저 정리한 뒤 구현 범위를 제안합니다.",
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-claude-code-dev",
		RuntimeKind:         RuntimeKindClaudeCode,
		ModelID:             "claude-sonnect-4-6",
		ModelLabel:          "Sonnect 4.6 (기본값)",
		WorkStatus:          AgentWorkStatusOffline,
		Editability:         AgentEditabilityEditable,
		AssignedTaskCount:   0,
		CreatedAt:           now.Add(-96 * time.Hour),
		UpdatedAt:           now.Add(-5 * time.Hour),
	}
}
