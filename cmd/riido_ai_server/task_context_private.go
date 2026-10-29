package main

import (
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func privateTaskContextReader(baseURL, graphqlURL string, timeout time.Duration) (riidoaiserver.AIAgentTaskContextReader, error) {
	readScope, err := envOptionalBool(envAssignmentReadScope)
	if err != nil {
		return nil, err
	}
	client, err := riidoaiserver.NewAIAgentPrivateTaskContextClient(riidoaiserver.AIAgentPrivateTaskContextClientConfig{
		AssignmentReadScope: readScope,
		BaseURL:             baseURL,
		GraphQLURL:          graphqlURL,
		Timeout:             timeout,
	})
	if err != nil {
		return nil, wrapEnvError(envTaskContextBaseURL, err)
	}
	return client, nil
}
