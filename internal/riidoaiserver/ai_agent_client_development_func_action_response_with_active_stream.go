package riidoaiserver

import (
	"strings"
)

func actionResponseWithActiveStream(response AIAgentTaskActionResponse, workspaceID string) AIAgentTaskActionResponse {
	if !taskThreadHasActiveStream(AIAgentTaskThreadRecord{AssignmentState: response.AssignmentState}) {
		return response
	}
	if strings.TrimSpace(response.TaskID) == "" || strings.TrimSpace(response.ThreadID) == "" || strings.TrimSpace(response.RunID) == "" {
		return response
	}
	link := activeStreamLinkForThread(AIAgentTaskThreadRecord{
		TaskID:   response.TaskID,
		ThreadID: response.ThreadID,
		RunID:    response.RunID,
	}, workspaceID)
	response.ActiveStream = &link
	return response
}
