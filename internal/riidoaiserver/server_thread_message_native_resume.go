package riidoaiserver

import (
	"context"
	"errors"
	"reflect"

	"github.com/teamswyg/riido-contracts/assignment"
)

var errNativeSessionUnavailable = errors.New("native session unavailable; start a new conversation explicitly")

// Select only the assignment behind the authorized thread. The daemon remains
// the authority for its retained device/runtime/workdir/session binding.
func (s Server) applyThreadNativeResume(ctx context.Context, principal AuthorizationResult, thread AIAgentTaskThreadRecord, req *AssignRequest, agentOwnerPrincipalID string) error {
	if thread.AssignmentID == "" || isNonDurableAssignmentID(thread.AssignmentID) || !nativeFollowupProvider(req.RuntimeProvider) {
		return nil
	}
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok {
		return assignmentDependencyFailure(errors.New("follow-up assignment source unavailable"))
	}
	projection, found, err := reader.LoadAssignmentProjection(ctx, thread.AssignmentID)
	if err != nil {
		return assignmentDependencyFailure(err)
	}
	if !found {
		return errNativeSessionUnavailable
	}
	previous := projection.Assignment
	resumeID := assignment.ResumeSessionIDForAssignment(previous)
	if !isTerminal(previous.State) {
		return nil
	}
	if resumeID == "" {
		return errNativeSessionUnavailable
	}
	if previous.ID != thread.AssignmentID || previous.TaskID != thread.TaskID || previous.AgentID != thread.AgentID || previous.WorkspaceID != principal.WorkspaceID || previous.WorkspaceID != req.WorkspaceID {
		return ErrAuthorizationForbidden
	}
	if previous.AgentID != req.AgentID || previous.RuntimeProvider != req.RuntimeProvider || previous.ModelID != req.ModelID || !reflect.DeepEqual(previous.Worktree, req.Worktree) {
		return ErrAIAgentTaskThreadConflict
	}
	if err := s.validateFollowupRuntimeBinding(thread, *req, agentOwnerPrincipalID); err != nil {
		return err
	}
	req.ResumeSessionID = resumeID
	return nil
}

func nativeFollowupProvider(provider string) bool {
	return provider == "codex" || provider == "claude" || provider == "cursor"
}

func (s Server) validateFollowupRuntimeBinding(thread AIAgentTaskThreadRecord, req AssignRequest, agentOwnerPrincipalID string) error {
	binding, runtime, err := s.resolveAgentRuntimeFact(req.AgentID)
	if err != nil {
		return assignmentDependencyFailure(err)
	}
	if binding.AgentID != req.AgentID || binding.RuntimeProvider != req.RuntimeProvider || runtimeProviderForAIAgentRuntime(runtime.Kind) != req.RuntimeProvider || binding.DeviceID == "" || binding.DaemonID == "" || binding.RuntimeID == "" || runtime.RuntimeID != binding.RuntimeID || runtime.DeviceID != binding.DeviceID || (runtime.DaemonID != "" && runtime.DaemonID != binding.DaemonID) || runtime.OwnerPrincipalID == "" {
		return ErrAIAgentTaskThreadConflict
	}
	// A connected account can own an Agent on another account's physical runtime.
	// Revalidate immutable Agent ownership against the authorized current profile.
	if snapshot := thread.AgentSnapshot; snapshot != nil && snapshot.OwnerPrincipalID != "" && snapshot.OwnerPrincipalID != agentOwnerPrincipalID {
		return ErrAuthorizationForbidden
	}
	return nil
}
