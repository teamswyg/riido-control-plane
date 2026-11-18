package riidoaiserver

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-contracts/assignment"
)

const workspaceRetryRequestPrefix = "workspace-retry-"

var errWorkspaceRetryAlreadyStarted = errors.New("a retry is already linked to this execution; use the latest execution in this conversation")

const workspaceRetryInputHelp = "reply with a linked PR number (PR #123), a branch in the same repository (branch: staging), or retry after restoring repository access"

// Both workspace setup and confirmed Codex preflight failures precede submission.
// A missing session ID by itself does not establish that it is safe to replay.
func (s Server) preparePreStartRetry(ctx context.Context, principal AuthorizationResult, thread AIAgentTaskThreadRecord, history AIAgentTaskThreadHistoryCollectionResponse, req *AssignRequest, taskContext *AIAgentTaskContext, body string) (bool, error) {
	if thread.AssignmentState != AgentAssignmentStateFailed {
		return false, nil
	}
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok {
		return false, nil
	}
	projection, found, err := reader.LoadAssignmentProjection(ctx, thread.AssignmentID)
	if err != nil {
		return false, assignmentDependencyFailure(err)
	}
	// Reconciliation can advance the thread before its diagnostic arrives.
	reportedWorkspaceFailure := thread.ExecutionContext != nil && thread.ExecutionContext.DiagnosticCode == workspacePreparationFailedDiagnostic
	workspaceFailure := reportedWorkspaceFailure || projection.WorkspacePreparationFailed
	if workspaceFailure && taskContext == nil {
		return false, nil
	}
	if !workspaceFailure && !confirmedCodexPreStartFailure(projection, req) {
		return false, nil
	}
	previous := projection.Assignment
	if !found || previous.ID != thread.AssignmentID || previous.TaskID != thread.TaskID ||
		previous.WorkspaceID != principal.WorkspaceID || previous.WorkspaceID != req.WorkspaceID ||
		previous.AgentID != thread.AgentID || previous.AgentID != req.AgentID ||
		previous.CreatedBy != principal.PrincipalID {
		return false, ErrAuthorizationForbidden
	}
	repositoryOptional := !workspaceFailure && previous.Worktree == nil && req.Worktree == nil
	if previous.State != AssignmentFailed || assignment.ResumeSessionIDForAssignment(previous) != "" ||
		(!repositoryOptional && !sameCheckoutRepository(previous.Worktree, req.Worktree)) {
		return false, nil
	}
	for _, row := range history.Threads {
		if row.TaskID == thread.TaskID && row.AgentID == thread.AgentID && row.ParentThreadID == thread.ThreadID {
			return false, errWorkspaceRetryAlreadyStarted
		}
	}
	instruction := taskMessageContextText(body, taskDocumentContextLimit)
	if workspaceFailure {
		instruction, err = workspaceRetryInstruction(instruction, req.Worktree, taskContext.linkedPullRequests)
		if err != nil {
			return false, err
		}
	}
	if instruction == "" {
		instruction = workspaceRetryOriginalInstruction(thread, history.Threads, principal.PrincipalID)
	}
	if instruction == "" {
		return false, errors.New("add the original instruction on the next line of your reply")
	}
	req.RequestID = workspaceRetryRequestPrefix + sha256Hex([]byte(previous.ID))
	req.Prompt = appendAIAgentTaskThreadMessagePrompt(req.Prompt, thread, CreateAIAgentTaskThreadMessageRequest{Body: instruction}, history.Threads...)
	return true, nil
}
