package riidoaiserver

import (
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const clientMessageNeedUserInput = "작업 내용을 확인했어요. 원하는 결과물이나 방향을 댓글로 알려주세요."

func applyAssignmentMetadataActionResponse(response *AIAgentTaskActionResponse, metadata map[string]string) {
	if !assignmentMetadataNeedsInput(metadata) {
		return
	}
	response.WorkStatus = AgentWorkStatusWaitingForUser
	response.AssignmentState = AgentAssignmentStateWaiting
	response.CommentKind = AgentTaskCommentNeedsInput
	ensureAssignmentResponseMessage(response, clientMessageNeedUserInput)
}

func assignmentMetadataNeedsInput(metadata map[string]string) bool {
	status := strings.TrimSpace(metadata[metadatakeys.AssignmentResultStatus.String()])
	switch strings.ToLower(strings.ReplaceAll(status, "-", "_")) {
	case "needs_input", "input_required", "requires_input", "waiting_for_user":
		return true
	default:
		return false
	}
}
