package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
)

var ErrDaemonProfileMismatch = errors.New("daemon profile mismatch")

func (s *DevelopmentAIAgentClientStore) ConfigureDaemonProfile(profile string) error {
	if s == nil {
		return nil
	}
	normalized, err := normalizeControlPlaneDaemonProfile(profile)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expectedDaemonProfile = normalized
	return nil
}

func normalizeControlPlaneDaemonProfile(profile string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "":
		return "", nil
	case "development":
		return "development", nil
	case "staging", "testnet":
		return "staging", nil
	case "production":
		return "production", nil
	default:
		return "", fmt.Errorf("unsupported daemon profile %q", profile)
	}
}

func daemonProfileMatches(expected, actual string) bool {
	expected, _ = normalizeControlPlaneDaemonProfile(expected)
	actual, _ = normalizeControlPlaneDaemonProfile(actual)
	return expected == "" || actual == expected
}

func daemonProfileMismatchError(expected, actual string) error {
	return fmt.Errorf("%w: expected %q, got %q",
		ErrDaemonProfileMismatch, expected, strings.TrimSpace(actual))
}
