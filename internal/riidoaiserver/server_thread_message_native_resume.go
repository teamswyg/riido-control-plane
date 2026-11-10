package riidoaiserver

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
)

var errNativeSessionUnavailable = errors.New("native session unavailable; start a new conversation explicitly")

// A terminal session can be incompatible even when no other thread is active.
var errNativeSessionConfigurationMismatch = errors.New("native session configuration does not match the requested agent, provider, model or worktree")
var errNativeSessionRuntimeMismatch = errors.New("native session runtime binding is inconsistent")

func isNativeSessionConflict(err error) bool {
	return errors.Is(err, errNativeSessionUnavailable) || errors.Is(err, errNativeSessionConfigurationMismatch) || errors.Is(err, errNativeSessionRuntimeMismatch)
}

// Select only the assignment behind the authorized thread. The daemon remains
// the authority for its retained device/runtime/workdir/session binding.
// Return the prior execution timestamp so shared-thread projection timestamps
// cannot cause already supplied peer context to be replayed.
func (s Server) applyThreadNativeResume(ctx context.Context, principal AuthorizationResult, thread AIAgentTaskThreadRecord, req *AssignRequest, agentOwnerPrincipalID string) (time.Time, error) {
	if thread.AssignmentID == "" || isNonDurableAssignmentID(thread.AssignmentID) || !nativeFollowupProvider(req.RuntimeProvider) {
		return time.Time{}, nil
	}
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok {
		return time.Time{}, assignmentDependencyFailure(errors.New("follow-up assignment source unavailable"))
	}
	projection, found, err := reader.LoadAssignmentProjection(ctx, thread.AssignmentID)
	if err != nil {
		return time.Time{}, assignmentDependencyFailure(err)
	}
	if !found {
		return time.Time{}, errNativeSessionUnavailable
	}
	previous := projection.Assignment
	resumeID := assignment.ResumeSessionIDForAssignment(previous)
	if !isTerminal(previous.State) {
		return time.Time{}, nil
	}
	if resumeID == "" {
		return time.Time{}, errNativeSessionUnavailable
	}
	if previous.ID != thread.AssignmentID || previous.TaskID != thread.TaskID || previous.AgentID != thread.AgentID || previous.WorkspaceID != principal.WorkspaceID || previous.WorkspaceID != req.WorkspaceID {
		return time.Time{}, ErrAuthorizationForbidden
	}
	worktreeChanged := previous.Worktree != nil && !reflect.DeepEqual(previous.Worktree, req.Worktree)
	if previous.AgentID != req.AgentID || previous.RuntimeProvider != req.RuntimeProvider || previous.ModelID != req.ModelID || worktreeChanged {
		return time.Time{}, errNativeSessionConfigurationMismatch
	}
	if err := s.validateFollowupRuntimeBinding(thread, *req, agentOwnerPrincipalID); err != nil {
		return time.Time{}, err
	}
	// Task context can discover a repository after a run starts, for example
	// when the Agent opens a PR. Keep the worktree bound to the native session.
	req.Worktree = cloneAssignmentWorktree(previous.Worktree)
	req.ResumeSessionID = resumeID
	if previous.State != AssignmentCompleted {
		return time.Time{}, nil // Delivery of the failed/interrupted input is unconfirmed.
	}
	return previous.CreatedAt, nil
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
		return errNativeSessionRuntimeMismatch
	}
	// A connected account can own an Agent on another account's physical runtime.
	// Revalidate immutable Agent ownership against the authorized current profile.
	if snapshot := thread.AgentSnapshot; snapshot != nil && snapshot.OwnerPrincipalID != "" && snapshot.OwnerPrincipalID != agentOwnerPrincipalID {
		return ErrAuthorizationForbidden
	}
	return nil
}
