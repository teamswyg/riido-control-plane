package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ensureOnboardingFixtureColorsLocked() {
	for i := range s.fixtures {
		if strings.TrimSpace(s.fixtures[i].TmpColor) != "" {
			continue
		}
		if color := aiAgentOnboardingFixtureTmpColors[s.fixtures[i].FixtureID]; color != "" {
			s.fixtures[i].TmpColor = color
		}
	}
}
