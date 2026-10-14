package main

import "testing"

func TestDaemonProfileFromEnvDefaultsAndNormalizes(t *testing.T) {
	clearRiidoAIServerEnv(t)
	got, err := daemonProfileFromEnv()
	if err != nil || got != "" {
		t.Fatalf("default daemon profile = %q, %v; want empty", got, err)
	}
	t.Setenv(envAIAgentDaemonProfile, " TESTNET ")
	got, err = daemonProfileFromEnv()
	if err != nil || got != "staging" {
		t.Fatalf("testnet daemon profile = %q, %v; want staging", got, err)
	}
	for _, want := range []string{"development", "production"} {
		t.Setenv(envAIAgentDaemonProfile, want)
		got, err = daemonProfileFromEnv()
		if err != nil || got != want {
			t.Fatalf("daemon profile = %q, %v; want %s", got, err, want)
		}
	}
}
