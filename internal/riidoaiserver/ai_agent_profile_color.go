package riidoaiserver

// AgentProfileColor is the fixed profile color of an AI Agent. It is chosen
// once when the agent is created (or backfilled when an older snapshot is
// loaded) and never changes afterwards.
type AgentProfileColor string

const (
	AgentProfileColorBlue   AgentProfileColor = "blue"
	AgentProfileColorOrange AgentProfileColor = "orange"
	AgentProfileColorGreen  AgentProfileColor = "green"
	AgentProfileColorPurple AgentProfileColor = "purple"
)

// agentProfileColorOrder is the canonical order used for least-used selection
// and for deterministic tie-breaks. It must stay in contract order.
var agentProfileColorOrder = []AgentProfileColor{
	AgentProfileColorBlue,
	AgentProfileColorOrange,
	AgentProfileColorGreen,
	AgentProfileColorPurple,
}

// onboardingAgentProfileColor is fixed for every agent created through the
// onboarding fixture path.
const onboardingAgentProfileColor = AgentProfileColorBlue

func isAgentProfileColor(color AgentProfileColor) bool {
	for _, candidate := range agentProfileColorOrder {
		if color == candidate {
			return true
		}
	}
	return false
}

// leastUsedAgentProfileColors returns the colors whose usage count is minimal,
// in canonical order. Colors missing from counts are treated as unused.
func leastUsedAgentProfileColors(counts map[AgentProfileColor]int) []AgentProfileColor {
	minCount := -1
	for _, color := range agentProfileColorOrder {
		if minCount < 0 || counts[color] < minCount {
			minCount = counts[color]
		}
	}
	candidates := make([]AgentProfileColor, 0, len(agentProfileColorOrder))
	for _, color := range agentProfileColorOrder {
		if counts[color] == minCount {
			candidates = append(candidates, color)
		}
	}
	return candidates
}

// fallbackAgentProfileColor maps an agent id to a color deterministically. It is
// used only when no stored color can be recovered for that agent.
func fallbackAgentProfileColor(agentID string) AgentProfileColor {
	return agentProfileColorOrder[agentIDHashIndex(agentID, len(agentProfileColorOrder))]
}

// profileColorOrFallback keeps a valid stored color and otherwise falls back to a
// deterministic color, so client-facing projections never emit an empty value.
func profileColorOrFallback(color AgentProfileColor, agentID string) AgentProfileColor {
	if isAgentProfileColor(color) {
		return color
	}
	return fallbackAgentProfileColor(agentID)
}
