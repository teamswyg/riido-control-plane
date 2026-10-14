package riidoaiserver

import (
	"net/http"
	"testing"
)

func assertStoppedThreads(t *testing.T, server http.Handler, base, token, taskID string, assignmentIDs ...string) {
	t.Helper()
	bytes := aiAgentSmokeRequest(t, server, http.MethodGet, base+"/tasks/"+taskID+"/threads", token, "", http.StatusOK)
	var threads AIAgentTaskThreadCollectionResponse
	aiAgentSmokeDecode(t, bytes, &threads)
	if threads.ActiveStream != nil {
		t.Fatalf("stopped agent should not expose active stream: %+v", threads.ActiveStream)
	}
	for _, assignmentID := range assignmentIDs {
		thread := taskThreadByAssignment(t, threads.Threads, assignmentID)
		if thread.AssignmentState != AgentAssignmentStateStopped || thread.WorkStatus != AgentWorkStatusIdle {
			t.Fatalf("thread %s not stopped: %+v", assignmentID, thread)
		}
	}
}
