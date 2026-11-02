package main

import (
	"os"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

// withAgentSourceState adds the optional monolith lookup that explains sources
// omitted from a follow-up capture. Both values empty leave it disabled.
func withAgentSourceState(config runtimeConfig, err error) (runtimeConfig, error) {
	if err != nil {
		return runtimeConfig{}, err
	}
	config.AgentSourceState = riidoaiserver.AgentSourceStateConfig{BaseURL: os.Getenv("RIIDO_AI_SERVER_AGENT_SOURCE_STATE_URL"), Key: os.Getenv("RIIDO_AI_SERVER_AGENT_SOURCE_STATE_KEY")}
	if err = config.AgentSourceState.Validate(); err != nil {
		return runtimeConfig{}, err
	}
	return config, nil
}
