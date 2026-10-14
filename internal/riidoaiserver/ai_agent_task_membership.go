package riidoaiserver

import (
	"sort"
	"time"
)

// Membership belongs to the task, independently of retained execution history.
// Removed records are tombstones: restoring a snapshot must not reassign an agent.
type AIAgentTaskMembership struct {
	AgentID       string                          `json:"agent_id"`
	AgentSnapshot *AIAgentTaskThreadAgentSnapshot `json:"agent_snapshot,omitempty"`
	AssignedAt    time.Time                       `json:"assigned_at"`
	RemovedAt     *time.Time                      `json:"removed_at,omitempty"`
}

type TaskAgentParticipantCollection struct {
	Agents []TaskAgentParticipantRecord `json:"agents"`
}

type TaskAgentParticipantRecord struct {
	AgentID             string `json:"agent_id"`
	Name                string `json:"name"`
	ProfileThumbnailURL string `json:"profile_thumbnail_url,omitempty"`
	IsOwnedByViewer     bool   `json:"is_owned_by_viewer"`
	IsStopRequired      bool   `json:"is_stop_required"`
	LatestAssignmentID  string `json:"latest_assignment_id,omitempty"`
}

func copyTaskAgentMemberships(source map[string]map[string]AIAgentTaskMembership) map[string]map[string]AIAgentTaskMembership {
	if source == nil {
		return nil
	}
	result := make(map[string]map[string]AIAgentTaskMembership, len(source))
	for taskID, members := range source {
		copied := make(map[string]AIAgentTaskMembership, len(members))
		for agentID, member := range members {
			member.AgentSnapshot = copyTaskThreadAgentSnapshot(member.AgentSnapshot)
			if member.RemovedAt != nil {
				removedAt := *member.RemovedAt
				member.RemovedAt = &removedAt
			}
			copied[agentID] = member
		}
		result[taskID] = copied
	}
	return result
}

// Legacy stopped/cancelled records cannot distinguish stop from explicit removal.
// Keep them absent rather than resurrect a participant the user removed.
func legacyTaskAgentMemberships(threads map[string][]AIAgentTaskThreadRecord) map[string]map[string]AIAgentTaskMembership {
	result := make(map[string]map[string]AIAgentTaskMembership)
	for taskID, history := range threads {
		seen := make(map[string]bool)
		for i := len(history) - 1; i >= 0; i-- {
			thread := history[i]
			if thread.AgentID == "" || seen[thread.AgentID] {
				continue
			}
			seen[thread.AgentID] = true
			if thread.AssignmentState == AgentAssignmentStateStopped || thread.AssignmentState == AgentAssignmentStateUnassigned {
				continue
			}
			if result[taskID] == nil {
				result[taskID] = make(map[string]AIAgentTaskMembership)
			}
			result[taskID][thread.AgentID] = AIAgentTaskMembership{AgentID: thread.AgentID, AgentSnapshot: copyTaskThreadAgentSnapshot(thread.AgentSnapshot), AssignedAt: thread.StartedAt}
		}
	}
	return result
}

func (s *DevelopmentAIAgentClientStore) assignTaskAgentMemberLocked(taskID string, agent AgentClientRecord) {
	if s.taskAgentMemberships == nil {
		s.taskAgentMemberships = legacyTaskAgentMemberships(s.taskThreads)
	}
	if s.taskAgentMemberships[taskID] == nil {
		s.taskAgentMemberships[taskID] = make(map[string]AIAgentTaskMembership)
	}
	current := s.taskAgentMemberships[taskID][agent.AgentID]
	now := time.Now().UTC()
	if current.AssignedAt.IsZero() || current.RemovedAt != nil {
		current.AssignedAt = now
	}
	current.AgentID = agent.AgentID
	current.AgentSnapshot = s.agentSnapshotFromAgent(agent, now)
	current.RemovedAt = nil
	s.taskAgentMemberships[taskID][agent.AgentID] = current
}

func (s *DevelopmentAIAgentClientStore) removeTaskAgentMemberLocked(taskID string, agent AgentClientRecord) {
	if s.taskAgentMemberships == nil {
		s.taskAgentMemberships = legacyTaskAgentMemberships(s.taskThreads)
	}
	if s.taskAgentMemberships[taskID] == nil {
		s.taskAgentMemberships[taskID] = make(map[string]AIAgentTaskMembership)
	}
	current := s.taskAgentMemberships[taskID][agent.AgentID]
	current.AgentID = agent.AgentID
	current.AgentSnapshot = s.agentSnapshotFromAgent(agent, time.Now().UTC())
	removedAt := time.Now().UTC()
	current.RemovedAt = &removedAt
	s.taskAgentMemberships[taskID][agent.AgentID] = current
}

func (s *DevelopmentAIAgentClientStore) taskAgentParticipantsLocked(principal AuthorizationResult) map[string]TaskAgentParticipantCollection {
	if s.taskAgentMemberships == nil {
		s.taskAgentMemberships = legacyTaskAgentMemberships(s.taskThreads)
	}
	result := make(map[string]TaskAgentParticipantCollection)
	for taskID, members := range s.taskAgentMemberships {
		ids := make([]string, 0, len(members))
		for id := range members {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, agentID := range ids {
			member := members[agentID]
			if member.RemovedAt != nil {
				continue
			}
			thread := AIAgentTaskThreadRecord{AgentID: agentID, AgentSnapshot: member.AgentSnapshot}
			if !s.taskThreadVisibleTo(principal, thread) {
				continue
			}
			snapshot := member.AgentSnapshot
			if agent, ok := s.agents[agentID]; ok {
				snapshot = s.agentSnapshotFromAgent(agent, member.AssignedAt)
			}
			if snapshot == nil {
				continue
			}
			participant := TaskAgentParticipantRecord{
				AgentID: agentID, Name: snapshot.Name, ProfileThumbnailURL: snapshot.ProfileThumbnailURL,
				IsOwnedByViewer: snapshot.OwnerPrincipalID == principal.PrincipalID,
			}
			if latest, ok := s.latestTaskThreadForAgentLocked(taskID, agentID); ok {
				participant.LatestAssignmentID = latest.AssignmentID
			}
			for _, run := range s.taskThreads[taskID] {
				if run.AgentID == agentID && taskThreadHasActiveStream(run) && run.AssignmentState != AgentAssignmentStateStopping {
					participant.IsStopRequired = true
				}
			}
			collection := result[taskID]
			collection.Agents = append(collection.Agents, participant)
			result[taskID] = collection
		}
	}
	return result
}
