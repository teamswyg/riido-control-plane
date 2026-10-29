package riidoaiserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AIAgentPrivateTaskContextClientConfig struct {
	AssignmentReadScope bool
	GraphQLURL          string
	BaseURL             string
	Timeout             time.Duration
	HTTPClient          *http.Client
}

type AIAgentPrivateTaskContextClient struct {
	assignmentReadScope bool
	readScopeEvidence   *taskReadEvidenceIndex
	graphqlURL          string
	projectionBundles   bool
	baseURL             string
	timeout             time.Duration
	httpClient          *http.Client
}

func NewAIAgentPrivateTaskContextClient(config AIAgentPrivateTaskContextClientConfig) (*AIAgentPrivateTaskContextClient, error) {
	baseURL, err := normalizeAIAgentTaskContextBaseURL(strings.TrimSpace(config.BaseURL))
	if err != nil {
		return nil, err
	}
	graphqlURL := baseURL
	if strings.TrimSpace(config.GraphQLURL) != "" {
		graphqlURL, err = normalizeAIAgentTaskContextBaseURL(strings.TrimSpace(config.GraphQLURL))
		if err != nil {
			return nil, fmt.Errorf("riidoaiserver: dashboard GraphQL URL: %w", err)
		}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultAIAgentTaskContextClientTimeout
	}
	if timeout < 0 {
		return nil, errors.New("riidoaiserver: private task context timeout must be positive")
	}
	return &AIAgentPrivateTaskContextClient{
		assignmentReadScope: config.AssignmentReadScope,
		graphqlURL:          graphqlURL,
		projectionBundles:   strings.TrimSpace(config.GraphQLURL) != "",
		baseURL:             baseURL,
		timeout:             timeout,
		httpClient:          aiAgentTaskContextHTTPClient(config.HTTPClient, timeout),
	}, nil
}
