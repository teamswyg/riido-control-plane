package main

import (
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func privateTaskContextReader(baseURL, graphqlURL string, timeout time.Duration) (riidoaiserver.AIAgentTaskContextReader, error) {
	client, err := riidoaiserver.NewAIAgentPrivateTaskContextClient(riidoaiserver.AIAgentPrivateTaskContextClientConfig{
		BaseURL:    baseURL,
		GraphQLURL: graphqlURL,
		Timeout:    timeout,
	})
	if err != nil {
		return nil, wrapEnvError(envTaskContextBaseURL, err)
	}
	return client, nil
}
