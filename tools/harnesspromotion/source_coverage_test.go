package main

import (
	"encoding/json"
	"os"
	"testing"
)

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

func TestHarnessPromotionChainBindsClaim(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(repoPath(root, "docs/30-architecture/evidence-graph.riido.json"))
	if err != nil {
		t.Fatal(err)
	}
	var graph struct {
		Chains []struct {
			ID     string   `json:"id"`
			Claims []string `json:"claims"`
		} `json:"chains"`
	}
	if err := json.Unmarshal(data, &graph); err != nil {
		t.Fatal(err)
	}
	for _, chain := range graph.Chains {
		if chain.ID == "harness_failure_promotion_loop" &&
			containsString(chain.Claims, "harness_promotion_must_run_after_failure") {
			return
		}
	}
	t.Fatal("harness failure promotion chain must bind its loop registry claim")
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
