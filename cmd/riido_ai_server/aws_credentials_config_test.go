package main

import (
	"strings"
	"testing"
)

func TestConfigFromEnvRejectsRemotePlainHTTPCredentialsEndpoint(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAIAgentClientDev, "true")
	t.Setenv(envAIAgentClientTable, "riido-ai-agent-development")
	t.Setenv(envAWSRegion, "ap-northeast-2")
	t.Setenv(envAWSContainerCredentialsFullURI, "http://metadata.example.com/credentials")
	t.Setenv(envAWSContainerAuthorizationToken, "Bearer metadata-token")

	if _, err := configFromEnv(); err == nil ||
		!strings.Contains(err.Error(), envAWSContainerCredentialsFullURI) ||
		!strings.Contains(err.Error(), "https") {
		t.Fatalf("expected remote plain HTTP credentials endpoint rejection, got %v", err)
	}
}
