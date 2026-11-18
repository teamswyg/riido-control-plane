package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func taskContextReaderFromEnv() (riidoaiserver.AIAgentTaskContextReader, error) {
	locatorVersion, err := taskContextLocatorVersionFromEnv()
	if err != nil {
		return nil, err
	}
	raw := taskContextRawEnv{
		baseURL:     strings.TrimSpace(os.Getenv(envTaskContextBaseURL)),
		graphqlURL:  strings.TrimSpace(os.Getenv(envTaskContextGraphQLURL)),
		workspaceID: strings.TrimSpace(os.Getenv(envTaskContextWorkspaceID)),
		teamID:      strings.TrimSpace(os.Getenv(envTaskContextTeamID)),
		apiKey:      strings.TrimSpace(os.Getenv(envTaskContextAPIKey)),
		timeoutRaw:  strings.TrimSpace(os.Getenv(envTaskContextTimeout)),
	}
	if raw.empty() {
		if locatorVersion == 2 {
			return nil, fmt.Errorf("%s requires private task context configuration", envTaskContextLocatorVersion)
		}
		return nil, nil
	}
	if raw.baseURL == "" {
		return nil, fmt.Errorf("%s is required when task context configuration is set", envTaskContextBaseURL)
	}
	timeout, err := envDurationSeconds(envTaskContextTimeout, 0)
	if err != nil {
		return nil, err
	}
	if raw.openAPIUnset() {
		return privateTaskContextReader(raw.baseURL, raw.graphqlURL, timeout)
	}
	if locatorVersion == 2 {
		return nil, fmt.Errorf("%s requires requester-private task context", envTaskContextLocatorVersion)
	}
	if !raw.openAPIComplete() {
		return nil, fmt.Errorf("%s, %s, and %s must be set together for OpenAPI task context; omit all three to use private JWT task context", envTaskContextWorkspaceID, envTaskContextTeamID, envTaskContextAPIKey)
	}
	client, err := riidoaiserver.NewAIAgentTaskContextClient(riidoaiserver.AIAgentTaskContextClientConfig{
		BaseURL:         raw.baseURL,
		WorkspaceID:     raw.workspaceID,
		TeamID:          raw.teamID,
		WorkspaceAPIKey: raw.apiKey,
		Timeout:         timeout,
	})
	if err != nil {
		return nil, wrapEnvError(envTaskContextBaseURL, err)
	}
	return client, nil
}
