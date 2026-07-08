package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/requirements"
)

func smokeFixtureManifest() manifest {
	return manifest{
		SchemaVersion:      requirements.ManifestSchema,
		ID:                 requirements.ExpectedID,
		Title:              "Smoke Matrix",
		RiidoTask:          "RIID-test",
		GeneratedDoc:       "docs/matrix.md",
		Workflow:           ".github/workflows/matrix.yml",
		EvidenceArtifact:   "matrix-evidence",
		OpenAPI:            "openapi.json",
		SmokeMatrix:        "smoke.json",
		SmokeSchemaVersion: "smoke.v1",
		OperationCounts:    operationCounts{Total: 2, V1: 1, V2: 1},
		RequiredEvidenceTests: []string{
			"TestHTTPAIAgentClientGeneratedEndpointSmokeV1",
			"TestHTTPAIAgentClientGeneratedEndpointSmokeV2",
		},
		SourceChecks: []sourceCheck{{
			Name:     "source",
			File:     "source.go",
			Contains: []string{"needle"},
		}},
		Loop: evidenceLoop{
			Observation:   "observed",
			Hypothesis:    "hypothesis",
			Execute:       "execute",
			Evaluate:      "evaluate",
			Retrospective: "retro",
		},
		NonGoals: []string{"live smoke"},
	}
}

func writeSmokeFixtureRepo(t *testing.T) (string, manifest) {
	t.Helper()
	repo := t.TempDir()
	m := smokeFixtureManifest()
	mustJSON(t, filepath.Join(repo, m.OpenAPI), map[string]any{
		"paths": map[string]any{
			"/v1/foo": map[string]any{
				"get": map[string]any{
					"operationId":    "getFoo",
					"x-riido-client": map[string]any{"generated_path": "v1.foo"},
				},
			},
			"/v2/bar": map[string]any{
				"post": map[string]any{
					"operationId":    "postBar",
					"x-riido-client": map[string]any{"generated_path": "v2.bar"},
				},
			},
		},
	})
	mustJSON(t, filepath.Join(repo, m.SmokeMatrix), smokeFixtureMatrix(m))
	if err := os.WriteFile(filepath.Join(repo, "source.go"), []byte("package x\n// needle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return repo, m
}
