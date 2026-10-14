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

func TestComposeAssignRequestWithTaskContextAddsWorktree(t *testing.T) {
	req, err := composeAssignRequestWithTaskContext("task-1", "component-1", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "placeholder",
	}, AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{
			ID:         "component-1",
			Title:      "Ship F3",
			BranchName: "RIID-4964-agent-profile-upload",
		},
		Document: AIAgentTaskContextDocument{Content: "Use real repo."},
		Repositories: []AIAgentTaskContextRepository{{
			FullName:      "teamswyg/riido-daemon",
			RepositoryURL: "https://github.com/teamswyg/riido-daemon",
			Source:        TaskContextRepositorySourceConnectedPullRequest,
		}},
	})
	if err != nil {
		t.Fatalf("compose assignment request: %v", err)
	}
	if req.Worktree == nil {
		t.Fatal("expected worktree")
	}
	if req.Worktree.RepositoryFullName != "teamswyg/riido-daemon" ||
		req.Worktree.RepositoryURL != "https://github.com/teamswyg/riido-daemon" ||
		req.Worktree.BranchName != "RIID-4964-agent-profile-upload" ||
		req.Worktree.Source != TaskContextRepositorySourceConnectedPullRequest {
		t.Fatalf("worktree = %+v", req.Worktree)
	}
}

func TestComposeAssignRequestDropsSensitiveRepositoryURL(t *testing.T) {
	req, err := composeAssignRequestWithTaskContext("task-1", "component-1", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "placeholder",
	}, AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{
			ID:         "component-1",
			Title:      "Ship F3 safely",
			BranchName: "RIID-4964-agent-profile-upload",
		},
		Document: AIAgentTaskContextDocument{Content: "Use real repo."},
		Repositories: []AIAgentTaskContextRepository{{
			FullName:      "teamswyg/riido-daemon",
			RepositoryURL: "https://github.com/teamswyg/riido-daemon?token=secret",
			Source:        TaskContextRepositorySourceConnectedPullRequest,
		}},
	})
	if err != nil {
		t.Fatalf("compose assignment request: %v", err)
	}
	if req.Worktree == nil {
		t.Fatal("expected worktree")
	}
	if req.Worktree.RepositoryFullName != "teamswyg/riido-daemon" || req.Worktree.RepositoryURL != "" {
		t.Fatalf("worktree should keep full_name and drop sensitive URL: %+v", req.Worktree)
	}
	for _, forbidden := range []string{"token", "secret"} {
		if strings.Contains(req.Prompt, forbidden) {
			t.Fatalf("prompt leaked sensitive repository URL component %q:\n%s", forbidden, req.Prompt)
		}
	}
	if !strings.Contains(req.Prompt, "- repository_url: not provided") {
		t.Fatalf("prompt should mark unsafe repository URL absent:\n%s", req.Prompt)
	}
}

func TestComposeAssignRequestDropsSensitiveRepositoryFullName(t *testing.T) {
	req, err := composeAssignRequestWithTaskContext("task-1", "component-1", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "placeholder",
	}, AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{
			ID:    "component-1",
			Title: "Ship F3 safely",
		},
		Document: AIAgentTaskContextDocument{Content: "Use real repo."},
		Repositories: []AIAgentTaskContextRepository{{
			FullName:      "teamswyg/riido-daemon?token=secret",
			RepositoryURL: "https://github.com/teamswyg/riido-daemon",
			Source:        TaskContextRepositorySourceConnectedPullRequest,
		}},
	})
	if err != nil {
		t.Fatalf("compose assignment request: %v", err)
	}
	if req.Worktree == nil {
		t.Fatal("expected worktree")
	}
	if req.Worktree.RepositoryFullName != "" || req.Worktree.RepositoryURL != "https://github.com/teamswyg/riido-daemon" {
		t.Fatalf("worktree should drop sensitive full_name and keep safe URL: %+v", req.Worktree)
	}
	for _, forbidden := range []string{"token", "secret"} {
		if strings.Contains(req.Prompt, forbidden) {
			t.Fatalf("prompt leaked sensitive repository full_name component %q:\n%s", forbidden, req.Prompt)
		}
	}
	if !strings.Contains(req.Prompt, "- full_name: not provided") {
		t.Fatalf("prompt should mark unsafe repository full_name absent:\n%s", req.Prompt)
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

func TestComposeAIAgentAssignmentPromptWithoutTaskContext(t *testing.T) {
	prompt, err := ComposeAIAgentAssignmentPromptWithoutTaskContext("task-smoke", "")
	if err != nil {
		t.Fatalf("compose prompt without task context: %v", err)
	}
	for _, want := range []string{
		"- task_id: task-smoke",
		"- component_id: task-smoke",
		"- title: task-smoke",
		"Task context was not available when this assignment was created.",
	} {
		if !strings.Contains(prompt.Prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt.Prompt)
		}
	}
	if prompt.HasRepository {
		t.Fatalf("fallback prompt should not select repository: %+v", prompt.SelectedRepository)
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
