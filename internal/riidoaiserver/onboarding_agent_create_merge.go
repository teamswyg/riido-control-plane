package riidoaiserver

import "sort"

// Preserve both the private receipt and its agent against ordinary cached
// writers. Deletion/editing observed by that writer remains a real mutation;
// conflicting mutations must reload rather than reviving an older profile.
func mergeCommittedOnboardingAgentCreates(baseline, committed AIAgentClientSnapshot, requested *AIAgentClientSnapshot) error {
	receipts, err := mergeOnboardingMap(baseline.OnboardingAgentCreates, requested.OnboardingAgentCreates, committed.OnboardingAgentCreates)
	if err != nil {
		return err
	}
	relevant := map[string]bool{}
	for _, set := range []map[string]onboardingAgentCreateReceipt{baseline.OnboardingAgentCreates, requested.OnboardingAgentCreates, committed.OnboardingAgentCreates} {
		for _, receipt := range set {
			relevant[receipt.AgentID] = true
		}
	}
	selectAgents := func(agents []AgentClientRecord) map[string]AgentClientRecord {
		result := map[string]AgentClientRecord{}
		for _, agent := range agents {
			if relevant[agent.AgentID] {
				result[agent.AgentID] = agent
			}
		}
		return result
	}
	agents, err := mergeOnboardingMap(selectAgents(baseline.Agents), selectAgents(requested.Agents), selectAgents(committed.Agents))
	if err != nil {
		return err
	}
	out := make([]AgentClientRecord, 0, len(requested.Agents)+len(agents))
	for _, agent := range requested.Agents {
		if !relevant[agent.AgentID] {
			out = append(out, agent)
		}
	}
	for _, agent := range agents {
		out = append(out, agent)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AgentID < out[j].AgentID })
	requested.Agents, requested.OnboardingAgentCreates = out, receipts
	return nil
}
