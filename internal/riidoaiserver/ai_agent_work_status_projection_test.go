package riidoaiserver

import "testing"

func TestProjectedAgentWorkStatusFromActiveThread(t *testing.T) {
	tests := []struct {
		name   string
		thread AIAgentTaskThreadRecord
		want   AgentWorkStatus
	}{
		{
			name:   "explicit waiting wins",
			thread: AIAgentTaskThreadRecord{WorkStatus: AgentWorkStatusWaitingForUser},
			want:   AgentWorkStatusWaitingForUser,
		},
		{
			name:   "queued assignment projects queued",
			thread: AIAgentTaskThreadRecord{AssignmentState: AgentAssignmentStateQueued},
			want:   AgentWorkStatusQueued,
		},
		{
			name:   "running assignment projects running",
			thread: AIAgentTaskThreadRecord{AssignmentState: AgentAssignmentStateRunning},
			want:   AgentWorkStatusRunning,
		},
		{
			name:   "unknown active projection defaults running",
			thread: AIAgentTaskThreadRecord{AssignmentState: AgentAssignmentStateCompleted},
			want:   AgentWorkStatusRunning,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := projectedAgentWorkStatusFromActiveThread(tt.thread); got != tt.want {
				t.Fatalf("projected status = %q, want %q", got, tt.want)
			}
		})
	}
}
