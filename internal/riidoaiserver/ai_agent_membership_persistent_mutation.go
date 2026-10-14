package riidoaiserver

import (
	"context"
	"errors"
	"time"
)

var errTaskAgentSnapshotChanged = errors.New("task Agent snapshot changed during membership commit")

func (s *PersistentAIAgentClientStore) forceReloadMembershipSnapshot(ctx context.Context) error {
	if owner, ok := s.snapshotStore.(interface{ requiresFreshTaskAgentMembership() bool }); !ok || !owner.requiresFreshTaskAgentMembership() {
		return s.reloadSnapshot(ctx)
	}
	s.snapshotReloadMu.Lock()
	s.snapshotLastReloadAt = time.Time{}
	s.snapshotReloadMu.Unlock()
	return s.reloadSnapshot(context.WithValue(ctx, taskAgentFreshSnapshotContextKey{}, true))
}

func (s *PersistentAIAgentClientStore) mutateTaskAgentMembership(ctx context.Context, taskID, agentID, publicationGeneration string, fence *taskAgentMembershipRemovalFence, publication bool, mutate func() (AIAgentTaskActionResponse, error)) (AIAgentTaskActionResponse, error) {
	for attempt := 0; attempt < 3; attempt++ {
		if err := s.forceReloadMembershipSnapshot(ctx); err != nil {
			return AIAgentTaskActionResponse{}, err
		}
		before, err := s.snapshot(s.snapshotNow().UTC())
		if err != nil {
			return AIAgentTaskActionResponse{}, err
		}
		if fence != nil && (before.TaskAgentMemberships[taskID][agentID].DurableGeneration != fence.durableGeneration || before.TaskAgentMemberships[taskID][agentID].IntentGateAssignmentID != fence.intentGateAssignmentID) {
			return AIAgentTaskActionResponse{}, ErrAIAgentMembershipChanged
		}
		response, err := mutate()
		if err != nil {
			return response, err
		}
		if agentID == "" {
			agentID = response.AgentID
		}
		baseline := before.TaskAgentMemberships[taskID][agentID]
		generation := baseline.DurableGeneration
		if publicationGeneration != "" {
			generation = publicationGeneration
		}
		if publication || publicationGeneration != "" {
			s.mu.Lock()
			member := s.taskAgentMemberships[taskID][agentID]
			member.DurableGeneration = generation
			if publication {
				member.IntentGateAssignmentID = ""
				if isIntentGateAssignmentID(response.AssignmentID) {
					member.IntentGateAssignmentID = response.AssignmentID
				}
			}
			s.taskAgentMemberships[taskID][agentID] = member
			s.mu.Unlock()
		}
		mutation := taskAgentMembershipSnapshotMutation{scope: taskAgentGenerationScope{taskID: taskID, agentID: agentID, generation: generation}, baseline: baseline}
		err = s.saveSnapshot(withTaskAgentSnapshotMutation(ctx, mutation))
		if err == nil {
			return response, nil
		}
		if restoreErr := s.restoreSnapshotPreservingSubscribers(before); restoreErr != nil {
			return response, restoreErr
		}
		if !errors.Is(err, errTaskAgentSnapshotChanged) {
			return response, err
		}
	}
	return AIAgentTaskActionResponse{}, ErrAIAgentMembershipChanged
}

type taskAgentMembershipObservationReader interface {
	observeTaskAgentMembershipRemoval(context.Context, AuthorizationResult, string, string) (AIAgentTaskThreadCollectionResponse, *taskAgentMembershipRemovalFence, error)
}
