package riidoaiserver

import (
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ConfigureDaemonClientCompatibility(
	policy DaemonClientCompatibilityPolicy,
) error {
	policy.MinimumVersion = strings.TrimSpace(policy.MinimumVersion)
	policy.LatestVersion = strings.TrimSpace(policy.LatestVersion)
	policy.DownloadURL = strings.TrimSpace(policy.DownloadURL)
	if _, ok := daemonVersionParts(policy.MinimumVersion); !ok {
		return errors.New("minimum daemon version must be semantic version")
	}
	if policy.LatestVersion != "" {
		if _, ok := daemonVersionParts(policy.LatestVersion); !ok {
			return errors.New("latest daemon version must be semantic version")
		}
	}
	s.mu.Lock()
	s.daemonClientPolicy = policy
	s.mu.Unlock()
	return nil
}
