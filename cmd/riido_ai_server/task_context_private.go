package main

import (
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func privateTaskContextReader(baseURL string, timeout time.Duration) (riidoaiserver.AIAgentTaskContextReader, error) {
	client, err := riidoaiserver.NewAIAgentPrivateTaskContextClient(riidoaiserver.AIAgentPrivateTaskContextClientConfig{
		BaseURL: baseURL,
		Timeout: timeout,
	})
	if err != nil {
		return nil, wrapEnvError(envTaskContextBaseURL, err)
	}
	return client, nil
}
