package riidoaiserver

import (
	"time"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func developmentPublicOpenClawAgent(now time.Time) AgentClientRecord {
	return AgentClientRecord{
		AgentID:             "agent-public-openclaw",
		OwnerPrincipalID:    "user-2",
		WorkspaceID:         defaultAIAgentClientWorkspaceID,
		Name:                "OpenClaw 공개 에이전트",
		ProfileThumbnailURL: "https://cdn.riido.io/dev/ai-agents/openclaw-public.png",
		Description:         "공개 워크스페이스 반복 작업 에이전트",
		Instruction:         "공개 워크스페이스에서 반복 가능한 보조 작업을 수행합니다.",
		Visibility:          AgentVisibilityPublic,
		RuntimeID:           "runtime-openclaw-shared",
		RuntimeKind:         RuntimeKindOpenClaw,
		ModelID:             providercatalog.DefaultOpenClawModelID,
		ModelLabel:          "OpenClaw 기본 모델",
		WorkStatus:          AgentWorkStatusIdle,
		Editability:         AgentEditabilityEditable,
		AssignedTaskCount:   0,
		CreatedAt:           now.Add(-48 * time.Hour),
		UpdatedAt:           now.Add(-4 * time.Hour),
	}
}

func developmentPrivateCursorAgent(now time.Time) AgentClientRecord {
	return AgentClientRecord{
		AgentID:             "agent-private-cursor",
		OwnerPrincipalID:    "user-2",
		WorkspaceID:         defaultAIAgentClientWorkspaceID,
		Name:                "Cursor 비공개 에이전트",
		ProfileThumbnailURL: "https://cdn.riido.io/dev/ai-agents/cursor-private.png",
		Description:         "소유자 전용 Cursor 코드 탐색 에이전트",
		Instruction:         "소유자 전용 Cursor 기반 코드 탐색을 수행합니다.",
		Visibility:          AgentVisibilityPrivate,
		RuntimeID:           "runtime-cursor-private",
		RuntimeKind:         RuntimeKindCursor,
		ModelID:             providercatalog.DefaultCursorModelID,
		ModelLabel:          "Cursor Auto",
		WorkStatus:          AgentWorkStatusIdle,
		Editability:         AgentEditabilityEditable,
		AssignedTaskCount:   0,
		CreatedAt:           now.Add(-24 * time.Hour),
		UpdatedAt:           now.Add(-3 * time.Hour),
	}
}
