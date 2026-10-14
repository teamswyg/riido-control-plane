package deploypolicy

import "testing"

func assertRuntimeCDRepoDocs(t *testing.T, readme, domain, migration string) {
	t.Helper()
	requireContains(t, readme, "RIID-4839")
	requireContains(t, readme, "RIID-4842")
	requireContains(t, readme, "runtime-cd-ownership.md")
	requireContains(t, domain, "RIID-4839")
	requireContains(t, domain, "RIID-4842")
	for _, want := range migrationRuntimeCDPhrases() {
		requireContains(t, migration, want)
	}
}

func migrationRuntimeCDPhrases() []string {
	return []string{
		"RIID-4839",
		"RIID-4842",
		"RIID-4844",
		"RIID-4845",
		"RIID-4853",
		"RIID-4855",
	}
}
