package riidoaiserver

func copyAgentOnboardingFixtures(fixtures []AgentOnboardingFixture) []AgentOnboardingFixture {
	return append([]AgentOnboardingFixture(nil), fixtures...)
}
