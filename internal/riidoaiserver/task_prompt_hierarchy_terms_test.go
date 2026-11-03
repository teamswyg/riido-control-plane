package riidoaiserver

import (
	"strings"
	"testing"
)

func TestAssignmentPromptMapsGoalToMilestone(t *testing.T) {
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-1",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{ID: "component-1", ComponentType: "task", Title: "KPI 설계 최종화"},
			Hierarchy: AIAgentTaskContextHierarchy{
				Project:   AIAgentTaskContextReference{Title: "뤼이도에 오신 것을 환영합니다"},
				Milestone: AIAgentTaskContextReference{Title: "기본 기능을 알려드릴게요"},
			},
		},
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}
	_, section, found := strings.Cut(prompt.Prompt, "## Hierarchy\n")
	if !found {
		t.Fatalf("prompt has no hierarchy section:\n%s", prompt.Prompt)
	}
	section, _, _ = strings.Cut(section, "\n\n")
	for _, want := range []string{"milestone = 목표 (상위 목표)", "project = 프로젝트", "- milestone: 기본 기능을 알려드릴게요"} {
		if !strings.Contains(section, want) {
			t.Fatalf("hierarchy section misses %q:\n%s", want, section)
		}
	}
}
