package main

import "testing"

func TestPromotionSourcesMustCoverEveryHarnessLoop(t *testing.T) {
	m := loadPromotionManifestForTest(t)
	registry := loadPromotionRegistryForTest(t, m)
	m.Sources = m.Sources[1:]
	if err := verifySourceCoverage(registry, m.Sources); err == nil {
		t.Fatal("expected missing promotion source to fail")
	}
}

func TestPromotionSourcesRejectDuplicateHarnessLoop(t *testing.T) {
	m := loadPromotionManifestForTest(t)
	registry := loadPromotionRegistryForTest(t, m)
	m.Sources = append(m.Sources, m.Sources[0])
	m.Sources[len(m.Sources)-1].ID = "duplicate"
	if err := verifySourceCoverage(registry, m.Sources); err == nil {
		t.Fatal("expected duplicate promotion source to fail")
	}
}

func TestPromotionSourcesRejectWorkflowDrift(t *testing.T) {
	m := loadPromotionManifestForTest(t)
	registry := loadPromotionRegistryForTest(t, m)
	m.Sources[0].SourceWorkflow = ".github/workflows/loop-registry.yml"
	if err := verifySourceCoverage(registry, m.Sources); err == nil {
		t.Fatal("expected workflow drift to fail")
	}
}

func loadPromotionRegistryForTest(t *testing.T, m manifest) loopRegistry {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := loadLoopRegistry(root, m.LoopRegistryManifest)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
