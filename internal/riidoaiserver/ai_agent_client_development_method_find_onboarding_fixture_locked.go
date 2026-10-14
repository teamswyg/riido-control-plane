package riidoaiserver

func (s *DevelopmentAIAgentClientStore) findOnboardingFixtureLocked(fixtureID string) (AgentOnboardingFixture, bool) {
	for _, fixture := range s.fixtures {
		if fixture.FixtureID == fixtureID {
			return fixture, true
		}
	}
	return AgentOnboardingFixture{}, false
}
