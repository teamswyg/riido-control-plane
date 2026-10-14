package main

import "testing"

func assertRequiredCategoriesCovered(t *testing.T, counts map[string]int) {
	t.Helper()
	for _, category := range loadManifestForTest(t).RequiredCategories {
		if counts[category] == 0 {
			t.Fatalf("category %s missing from evidence counts: %+v", category, counts)
		}
	}
}
