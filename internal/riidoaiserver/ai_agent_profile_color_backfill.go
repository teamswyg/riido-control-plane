package riidoaiserver

import (
	"hash/fnv"
	"math/rand/v2"
	"sort"
	"strings"
)

// backfillAgentProfileColors gives every agent without a valid profile color one,
// using only snapshot content so that two instances loading the same snapshot
// compute the same colors. Agents with a tmp_color are fixture-created onboarding
// agents and get blue. The remaining agents take a least-used color within their
// workspace scope, processed in created_at and agent_id order.
func backfillAgentProfileColors(agents map[string]AgentClientRecord, scopeOf func(AgentClientRecord) string) {
	counts := map[string]map[AgentProfileColor]int{}
	var pending []AgentClientRecord
	for _, agent := range agents {
		scope := scopeOf(agent)
		switch {
		case isAgentProfileColor(agent.ProfileColor):
			addAgentProfileColorCount(counts, scope, agent.ProfileColor)
		case strings.TrimSpace(agent.TmpColor) != "":
			agent.ProfileColor = onboardingAgentProfileColor
			agents[agent.AgentID] = agent
			addAgentProfileColorCount(counts, scope, agent.ProfileColor)
		default:
			pending = append(pending, agent)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		if !pending[i].CreatedAt.Equal(pending[j].CreatedAt) {
			return pending[i].CreatedAt.Before(pending[j].CreatedAt)
		}
		return pending[i].AgentID < pending[j].AgentID
	})
	for _, agent := range pending {
		scope := scopeOf(agent)
		candidates := leastUsedAgentProfileColors(counts[scope])
		agent.ProfileColor = candidates[agentIDHashIndex(agent.AgentID, len(candidates))]
		agents[agent.AgentID] = agent
		addAgentProfileColorCount(counts, scope, agent.ProfileColor)
	}
}

// backfillTaskSnapshotProfileColors fills the profile color of task thread and
// membership snapshots written before profile colors existed. A snapshot takes the
// color of its current agent record; a snapshot whose agent is gone falls back to
// a deterministic color derived from the agent id.
func backfillTaskSnapshotProfileColors(
	threads map[string][]AIAgentTaskThreadRecord,
	memberships map[string]map[string]AIAgentTaskMembership,
	agents map[string]AgentClientRecord,
) {
	for taskID := range threads {
		for i := range threads[taskID] {
			fillTaskSnapshotProfileColor(threads[taskID][i].AgentSnapshot, agents)
		}
	}
	for _, members := range memberships {
		for _, member := range members {
			fillTaskSnapshotProfileColor(member.AgentSnapshot, agents)
		}
	}
}

func fillTaskSnapshotProfileColor(snapshot *AIAgentTaskThreadAgentSnapshot, agents map[string]AgentClientRecord) {
	if snapshot == nil || isAgentProfileColor(snapshot.ProfileColor) {
		return
	}
	if agent, ok := agents[snapshot.AgentID]; ok && isAgentProfileColor(agent.ProfileColor) {
		snapshot.ProfileColor = agent.ProfileColor
		return
	}
	snapshot.ProfileColor = fallbackAgentProfileColor(snapshot.AgentID)
}

func (s *DevelopmentAIAgentClientStore) ensureAgentProfileColorsLocked() {
	backfillAgentProfileColors(s.agents, s.agentWorkspaceID)
	backfillTaskSnapshotProfileColors(s.taskThreads, s.taskAgentMemberships, s.agents)
}

// chooseNewAgentProfileColorLocked picks one of the least-used colors in the
// workspace. The caller must hold s.mu so the count and the new agent are
// observed together.
func (s *DevelopmentAIAgentClientStore) chooseNewAgentProfileColorLocked(workspaceID string) AgentProfileColor {
	counts := map[AgentProfileColor]int{}
	for _, agent := range s.agents {
		if s.agentWorkspaceID(agent) == workspaceID && isAgentProfileColor(agent.ProfileColor) {
			counts[agent.ProfileColor]++
		}
	}
	candidates := leastUsedAgentProfileColors(counts)
	return candidates[s.profileColorIndexLocked(len(candidates))]
}

func (s *DevelopmentAIAgentClientStore) profileColorIndexLocked(n int) int {
	if n <= 1 {
		return 0
	}
	index := rand.IntN(n)
	if s.profileColorIndex != nil {
		index = s.profileColorIndex(n)
	}
	if index < 0 || index >= n {
		return 0
	}
	return index
}

func addAgentProfileColorCount(counts map[string]map[AgentProfileColor]int, scope string, color AgentProfileColor) {
	if counts[scope] == nil {
		counts[scope] = map[AgentProfileColor]int{}
	}
	counts[scope][color]++
}

func agentIDHashIndex(agentID string, n int) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(strings.TrimSpace(agentID)))
	return int(hash.Sum32() % uint32(n))
}
