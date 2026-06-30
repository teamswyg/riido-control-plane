package riidoaiserver

import (
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const clientMessageNeedUserInput = "어떤 작업부터 진행할까요?"

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
