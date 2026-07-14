package main

import (
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func daemonClientCompatibilityPolicyFromEnv() riidoaiserver.DaemonClientCompatibilityPolicy {
	policy := riidoaiserver.DefaultDaemonClientCompatibilityPolicy()
	if value := strings.TrimSpace(os.Getenv(envMinimumDaemonVersion)); value != "" {
		policy.MinimumVersion = value
	}
	if value := strings.TrimSpace(os.Getenv(envLatestDaemonVersion)); value != "" {
		policy.LatestVersion = value
	}
	if value := strings.TrimSpace(os.Getenv(envDaemonDownloadURL)); value != "" {
		policy.DownloadURL = value
	}
	return policy
}
