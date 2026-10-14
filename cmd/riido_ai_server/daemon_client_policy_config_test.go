package main

import "testing"

func TestDaemonClientCompatibilityPolicyFromEnv(t *testing.T) {
	t.Setenv(envMinimumDaemonVersion, "v0.0.60")
	t.Setenv(envLatestDaemonVersion, "v0.0.70")
	t.Setenv(envDaemonDownloadURL, "https://download.example/riido")
	policy := daemonClientCompatibilityPolicyFromEnv()
	if policy.MinimumVersion != "v0.0.60" || policy.LatestVersion != "v0.0.70" ||
		policy.DownloadURL != "https://download.example/riido" {
		t.Fatalf("policy = %+v", policy)
	}
}
