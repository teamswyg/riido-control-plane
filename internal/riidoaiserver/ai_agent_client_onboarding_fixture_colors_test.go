package riidoaiserver

import "testing"

func TestDevelopmentStoreEnsureOnboardingFixtureColorsLocked(t *testing.T) {
	store := &DevelopmentAIAgentClientStore{
		fixtures: []AgentOnboardingFixture{
			{
				FixtureID: "riido_pm",
				TmpColor:  "   ",
			},
			{
				FixtureID: "custom_fixture",
			},
			{
				FixtureID: "hongdo_frontend",
				TmpColor:  "#KEEP",
			},
		},
	}

	store.ensureOnboardingFixtureColorsLocked()

	if got, want := store.fixtures[0].TmpColor, "#C9A452"; got != want {
		t.Fatalf("blank mapped fixture color = %q, want %q", got, want)
	}
	if got := store.fixtures[1].TmpColor; got != "" {
		t.Fatalf("unknown fixture color = %q, want empty", got)
	}
	if got, want := store.fixtures[2].TmpColor, "#KEEP"; got != want {
		t.Fatalf("existing fixture color = %q, want %q", got, want)
	}
}
