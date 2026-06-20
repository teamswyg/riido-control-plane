package riidoaiserver

import "strings"

func composeAssignRequestWithTaskContext(taskID, componentID string, req AssignRequest, contextSnapshot AIAgentTaskContext) (AssignRequest, error) {
	composed, err := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{
		TaskID:  taskID,
		Context: contextSnapshot,
	})
	if err != nil {
		return AssignRequest{}, err
	}
	req.Prompt = composed.Prompt
	if composed.HasRepository {
		req.Worktree = assignmentWorktreeFromTaskContext(contextSnapshot, composed.SelectedRepository)
	}
	if strings.TrimSpace(req.ComponentID) == "" {
		req.ComponentID = strings.TrimSpace(contextSnapshot.Component.ID)
		if req.ComponentID == "" {
			req.ComponentID = componentID
		}
	}
	return req, nil
}

func composeAssignRequestWithoutTaskContext(taskID, componentID string, req AssignRequest) (AssignRequest, error) {
	composed, err := ComposeAIAgentAssignmentPromptWithoutTaskContext(taskID, componentID)
	if err != nil {
		return AssignRequest{}, err
	}
	req.Prompt = composed.Prompt
	if strings.TrimSpace(req.ComponentID) == "" {
		req.ComponentID = strings.TrimSpace(componentID)
	}
	return req, nil
}

func assignmentWorktreeFromTaskContext(contextSnapshot AIAgentTaskContext, repository AIAgentTaskContextRepository) *AssignmentWorktree {
	worktree := &AssignmentWorktree{
		RepositoryFullName: safeAIAgentRepositoryFullName(repository.FullName),
		RepositoryURL:      safeAIAgentRepositoryURL(repository.RepositoryURL),
		BranchName:         strings.TrimSpace(contextSnapshot.Component.BranchName),
		IsPrivate:          repository.IsPrivate,
		Source:             strings.TrimSpace(repository.Source),
	}
	if worktree.RepositoryFullName == "" && worktree.RepositoryURL == "" {
		return nil
	}
	return worktree
}
