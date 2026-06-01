package riidoaiserver

import (
	"strings"
	"testing"
)

func TestComposeAIAgentAssignmentPromptUsesTaskContextSnapshot(t *testing.T) {
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-1",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{
				ID:            "component-1",
				ComponentType: "task",
				Title:         "Ship AI Agent assignment",
				KeyNumber:     "4799",
				BranchName:    "RIID-4799-assignment-prompt-composer",
			},
			Document: AIAgentTaskContextDocument{
				Content:       "Implement the prompt composer and keep the client assign request small.",
				ContentFormat: "markdown",
			},
			Hierarchy: AIAgentTaskContextHierarchy{
				Project:   AIAgentTaskContextReference{KeyNumber: "4539", Title: "[v1.22] AI Contributors"},
				Milestone: AIAgentTaskContextReference{KeyNumber: "4719", Title: "Riido AI Agent Policy"},
			},
			Repositories: []AIAgentTaskContextRepository{
				{
					ID:       "repo-b",
					FullName: "teamswyg/riido-daemon",
					Source:   TaskContextRepositorySourceWorkspaceConnectedRepo,
				},
				{
					ID:            "repo-a",
					FullName:      "teamswyg/riido-control-plane",
					RepositoryURL: "https://github.com/teamswyg/riido-control-plane",
					Source:        TaskContextRepositorySourceConnectedPullRequest,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}
	if !prompt.HasRepository {
		t.Fatalf("expected selected repository")
	}
	if prompt.SelectedRepository.FullName != "teamswyg/riido-control-plane" {
		t.Fatalf("selected repository = %+v", prompt.SelectedRepository)
	}
	wantParts := []string{
		"# Riido AI Agent Assignment",
		"- task_id: task-1",
		"- component_id: component-1",
		"- key_number: 4799",
		"- title: Ship AI Agent assignment",
		"- branch_name: RIID-4799-assignment-prompt-composer",
		"- full_name: teamswyg/riido-control-plane",
		"- source: connected_pull_request",
		"- project: RIID-4539 [v1.22] AI Contributors",
		"- milestone: RIID-4719 Riido AI Agent Policy",
		"- content_format: markdown",
		"Implement the prompt composer and keep the client assign request small.",
	}
	for _, want := range wantParts {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}

func TestComposeAIAgentAssignmentPromptFallsBackWithoutRepository(t *testing.T) {
	prompt, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID: "task-2",
		Context: AIAgentTaskContext{
			Component: AIAgentTaskContextComponent{Title: "No repository task"},
		},
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}
	if prompt.HasRepository {
		t.Fatalf("repository should be absent: %+v", prompt.SelectedRepository)
	}
	for _, want := range []string{
		"- component_id: task-2",
		"- full_name: not provided",
		"not provided",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
}

func TestComposeAIAgentAssignmentPromptRejectsEmptyContext(t *testing.T) {
	_, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task-3"})
	if err == nil || !strings.Contains(err.Error(), "title or document content") {
		t.Fatalf("expected empty context error, got %v", err)
	}
}

func TestSelectAIAgentTaskContextRepositoryOrdersDeterministically(t *testing.T) {
	got, ok := SelectAIAgentTaskContextRepository([]AIAgentTaskContextRepository{
		{ID: "z", FullName: "teamswyg/z", Source: "unknown"},
		{ID: "b", FullName: "teamswyg/b", Source: TaskContextRepositorySourceWorkspaceConnectedRepo},
		{ID: "a", FullName: "teamswyg/a", Source: TaskContextRepositorySourceWorkspaceConnectedRepo},
	})
	if !ok {
		t.Fatalf("expected repository")
	}
	if got.ID != "a" {
		t.Fatalf("selected repository = %+v", got)
	}
}
