package riidoaiserver

import "testing"

func TestAssignmentStartupQueueDoesNotAssertAnotherTaskIsRunning(t *testing.T) {
	for _, state := range []AssignmentState{AssignmentQueued, AssignmentLeased} {
		response := AIAgentTaskActionResponse{}
		applyAssignmentStateActionResponse(&response, state)
		if response.Message == clientMessageAgentBusyQueued || response.Message == "" {
			t.Fatalf("%s must have a neutral startup message: %q", state, response.Message)
		}
	}
	if got := clientVisibleTaskThreadText("agent assignment is queued"); got == clientMessageAgentBusyQueued {
		t.Fatal("generic queued runtime text incorrectly claims another task is running")
	}
}

func TestAssignmentStartReplacesQueueStatusButPreservesAgentOutput(t *testing.T) {
	for _, state := range []AssignmentState{AssignmentReady, AssignmentRunning} {
		for _, message := range []string{clientMessageAgentBusyQueued, "agent assignment is queued"} {
			response := assignmentEventActionResponse(AIAgentTaskThreadRecord{}, state, message, nil)
			if response.Message != clientMessageTaskRunning {
				t.Fatalf("%s retained queue text: %q", state, response.Message)
			}
		}
		for _, message := range []string{"Inspecting the requested files", clientMessageAgentBusyQueued + " quoted in a report"} {
			response := assignmentEventActionResponse(AIAgentTaskThreadRecord{}, state, message, nil)
			if response.Message != message {
				t.Fatalf("%s lost agent output: %q", state, response.Message)
			}
		}
	}
}

func TestRunningSnapshotDoesNotRestoreLegacyQueueStatus(t *testing.T) {
	for _, message := range []string{clientMessageAgentBusyQueued, "agent assignment is queued"} {
		thread := AIAgentTaskThreadRecord{
			WorkStatus: AgentWorkStatusRunning, AssignmentState: AgentAssignmentStateRunning,
			CommentKind: AgentTaskCommentRuntimeProgress, Message: message,
		}
		for _, useLineFallback := range []bool{false, true} {
			if useLineFallback {
				thread.Message = ""
				thread.Lines = []AgentThreadProgressLine{{Seq: 1, Message: message}}
			}
			if got := clientVisibleTaskThreadMessage(thread); got != clientMessageTaskRunning {
				t.Fatalf("running snapshot retained queue status (line=%v): %q", useLineFallback, got)
			}
		}
	}
}
