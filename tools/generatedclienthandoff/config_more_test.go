package main

import (
	"strings"
	"testing"
)

func TestNormalizeConfigFillsSourceRefAndGeneratedAt(t *testing.T) {
	t.Parallel()
	cfg := minimalValidConfig()
	got, err := normalizeConfig(cfg)
	if err != nil {
		t.Fatalf("normalizeConfig: %v", err)
	}
	if got.SourceRef != cfg.SourceCommit {
		t.Fatalf("source ref = %q, want commit", got.SourceRef)
	}
	if len(got.GeneratedAt) != len("2006-01-02") {
		t.Fatalf("generated at = %q, want yyyy-mm-dd", got.GeneratedAt)
	}
}

func TestNormalizeConfigRejectsMissingIdentity(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(*config){
		"source-commit": func(cfg *config) { cfg.SourceCommit = " " },
		"target-branch": func(cfg *config) { cfg.TargetBranch = "" },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := minimalValidConfig()
			mutate(&cfg)
			_, err := normalizeConfig(cfg)
			if err == nil || !strings.Contains(err.Error(), name+" is required") {
				t.Fatalf("error = %v, want missing %s", err, name)
			}
		})
	}
}

func minimalValidConfig() config {
	return config{
		OpenAPI:      "openapi.json",
		DSL:          "contract.dsl.json",
		IR:           "contract.ir.json",
		Core:         "aiAgentClient.ts",
		React:        "aiAgentClient.react.ts",
		Out:          "out",
		SourceCommit: "abcdef",
		TargetBranch: "RIID-100-generated-handoff",
	}
}
