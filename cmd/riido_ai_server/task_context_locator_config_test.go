package main

import "testing"

func TestTaskContextLocatorVersionConfig(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want int
	}{{"", 1}, {"1", 1}, {"2", 2}, {"3", 0}, {"native", 0}} {
		t.Setenv(envTaskContextLocatorVersion, tc.raw)
		got, err := taskContextLocatorVersionFromEnv()
		if got != tc.want || (err != nil) != (tc.want == 0) {
			t.Fatalf("locator %q: %d %v", tc.raw, got, err)
		}
	}
}

func TestTaskContextNativeRequiresPrivateConfiguration(t *testing.T) {
	t.Setenv(envTaskContextLocatorVersion, "2")
	for _, key := range []string{envTaskContextBaseURL, envTaskContextGraphQLURL, envTaskContextWorkspaceID, envTaskContextTeamID, envTaskContextAPIKey, envTaskContextTimeout} {
		t.Setenv(key, "")
	}
	if _, err := taskContextReaderFromEnv(); err == nil {
		t.Fatal("unconfigured native reader accepted")
	}
	t.Setenv(envTaskContextBaseURL, "https://example.test")
	t.Setenv(envTaskContextWorkspaceID, "workspace")
	t.Setenv(envTaskContextTeamID, "team")
	t.Setenv(envTaskContextAPIKey, "key")
	if _, err := taskContextReaderFromEnv(); err == nil {
		t.Fatal("native OpenAPI reader accepted")
	}
}
