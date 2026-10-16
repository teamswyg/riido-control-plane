package riidoaiserver

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// errOnboardingCompletionIneligible rejects agents that were not created through
// the onboarding fixture path. It maps to the default 400 response.
var errOnboardingCompletionIneligible = errors.New("onboarding completion requires an agent created through onboarding")

// errOnboardingCompletionConflict maps to 409. It covers one-time completion,
// prior task history, active executions, and request_id reuse for another target.
var errOnboardingCompletionConflict = errors.New("onboarding completion conflicts with existing agent history")

// errOnboardingCompletionStopConflict rejects a stop that names, or resolves to, an
// onboarding-seed record. That record is a completed message, not an execution, so
// the stop must not overwrite it. It maps to 409.
var errOnboardingCompletionStopConflict = errors.New("onboarding completion is a completed record and cannot be stopped")

// onboardingCompletionReceiptKey scopes request_id to workspace, owner principal and
// operation. It never matches assignment receipts, which use identifiedAssignmentID.
func onboardingCompletionReceiptKey(principal AuthorizationResult, requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("onboarding-completion\x00" + principal.WorkspaceID + "\x00" + principal.PrincipalID + "\x00" + requestID))
	return fmt.Sprintf("onboarding-completion-%x", sum)
}

// checkOnboardingCompletionEligibleLocked runs after the idempotent replay checks.
// The owner check is the caller's responsibility.
func (s *DevelopmentAIAgentClientStore) checkOnboardingCompletionEligibleLocked(taskID, agentID string) error {
	if !s.onboardingCreated[agentID] {
		return errOnboardingCompletionIneligible
	}
	if _, done := s.onboardingCompletions[agentID]; done {
		return fmt.Errorf("%w: agent already completed onboarding", errOnboardingCompletionConflict)
	}
	if _, active := s.activeTaskThreadForAgentLocked(taskID, agentID); active {
		return fmt.Errorf("%w: agent already has an active execution on this task", errOnboardingCompletionConflict)
	}
	if _, member := s.taskAgentMemberships[taskID][agentID]; member {
		return fmt.Errorf("%w: task already has this agent as a participant", errOnboardingCompletionConflict)
	}
	for _, thread := range s.taskThreads[taskID] {
		if thread.AgentID == agentID {
			return fmt.Errorf("%w: task already has assignment history for this agent", errOnboardingCompletionConflict)
		}
	}
	return nil
}

func (s *DevelopmentAIAgentClientStore) markOnboardingCreatedAgentLocked(agentID string) {
	if s.onboardingCreated == nil {
		s.onboardingCreated = map[string]bool{}
	}
	s.onboardingCreated[agentID] = true
}

func (s *DevelopmentAIAgentClientStore) markOnboardingCompletedLocked(agentID, taskID string, thread AIAgentTaskThreadRecord) {
	if s.onboardingCompletions == nil {
		s.onboardingCompletions = map[string]string{}
	}
	if s.onboardingCompletionThreads == nil {
		s.onboardingCompletionThreads = map[string]AIAgentTaskThreadRecord{}
	}
	s.onboardingCompletions[agentID] = taskID
	s.onboardingCompletionThreads[agentID] = copyTaskThread(thread)
}

// onboardingCompletionThreadLocked returns the completion thread for this agent and
// task. The completion record gates every lookup: a thread without a record for this
// task, such as one left by a deleted agent with the same ID, never matches. With a
// record, it reads the live task history first and falls back to the copy kept with
// the record once truncation has dropped the live thread.
func (s *DevelopmentAIAgentClientStore) onboardingCompletionThreadLocked(taskID, agentID string) (AIAgentTaskThreadRecord, bool) {
	if s.onboardingCompletions[agentID] != taskID {
		return AIAgentTaskThreadRecord{}, false
	}
	assignmentID := onboardingSeedAssignmentID(taskID, agentID)
	if thread, ok := s.onboardingSeedThreadLocked(taskID, assignmentID); ok {
		return thread, true
	}
	thread, ok := s.onboardingCompletionThreads[agentID]
	if !ok || thread.TaskID != taskID || thread.AssignmentID != assignmentID {
		return AIAgentTaskThreadRecord{}, false
	}
	return copyTaskThread(thread), true
}

func onboardingCreatedAgentIDsSnapshot(agents map[string]bool) []string {
	ids := make([]string, 0, len(agents))
	for id, created := range agents {
		if created {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return ids
}

func onboardingCreatedAgentsFromSnapshot(ids []string) map[string]bool {
	agents := make(map[string]bool, len(ids))
	for _, id := range ids {
		agents[strings.TrimSpace(id)] = true
	}
	return agents
}

// backfillOnboardingCreatedFromFixtureTmpColor marks agents whose tmp_color is one of
// the onboarding fixture colors. Only the fixture-create path writes tmp_color: the
// public create request has no tmp_color field, and the normal create path passes an
// empty value. So this check proves fixture provenance for snapshots that predate
// onboarding_created_agent_ids. Any other agent is left unmarked.
func backfillOnboardingCreatedFromFixtureTmpColor(created map[string]bool, agents map[string]AgentClientRecord) {
	fixtureColors := make(map[string]bool, len(aiAgentOnboardingFixtureTmpColors))
	for _, color := range aiAgentOnboardingFixtureTmpColors {
		fixtureColors[color] = true
	}
	for agentID, agent := range agents {
		if fixtureColors[strings.TrimSpace(agent.TmpColor)] {
			created[agentID] = true
		}
	}
}

func cloneOnboardingCompletedTasks(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for agentID, taskID := range input {
		result[agentID] = taskID
	}
	return result
}

func cloneOnboardingCompletionThreads(input map[string]AIAgentTaskThreadRecord) map[string]AIAgentTaskThreadRecord {
	result := make(map[string]AIAgentTaskThreadRecord, len(input))
	for agentID, thread := range input {
		result[agentID] = copyTaskThread(thread)
	}
	return result
}

func errOnboardingCompletionConflictForRequestID() error {
	return fmt.Errorf("%w: request_id was already used for another task or agent", errOnboardingCompletionConflict)
}
