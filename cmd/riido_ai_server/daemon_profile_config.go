package main

import (
	"fmt"
	"os"
	"strings"
)

func daemonProfileFromEnv() (string, error) {
	profile := strings.ToLower(strings.TrimSpace(os.Getenv(envAIAgentDaemonProfile)))
	switch profile {
	case "":
		return "", nil
	case "development":
		return "development", nil
	case "staging", "testnet":
		return "staging", nil
	case "production":
		return "production", nil
	default:
		return "", fmt.Errorf("%s must be one of development, staging, testnet, production", envAIAgentDaemonProfile)
	}
}
