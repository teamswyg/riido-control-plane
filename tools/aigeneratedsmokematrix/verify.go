package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/requirements"
)

type verifyResult struct {
	Counts operationCounts
}

func verifyAll(repo string, m manifest) (verifyResult, error) {
	if err := verifyManifestShape(m); err != nil {
		return verifyResult{}, err
	}
	ops, counts, err := loadOpenAPIGenerated(pathutil.Resolve(repo, m.OpenAPI))
	if err != nil {
		return verifyResult{}, fmt.Errorf("load OpenAPI: %w", err)
	}
	matrix, err := loadSmokeMatrix(pathutil.Resolve(repo, m.SmokeMatrix))
	if err != nil {
		return verifyResult{}, fmt.Errorf("load smoke matrix: %w", err)
	}
	if err := verifyCounts(m.OperationCounts, counts); err != nil {
		return verifyResult{}, err
	}
	if err := verifyMatrix(m, ops, matrix); err != nil {
		return verifyResult{}, err
	}
	if err := verifySourceChecks(repo, m.SourceChecks); err != nil {
		return verifyResult{}, err
	}
	return verifyResult{Counts: counts}, verifyLoop(m.Loop)
}

func verifyManifestShape(m manifest) error {
	if m.SchemaVersion != requirements.ManifestSchema || m.ID != requirements.ExpectedID {
		return fmt.Errorf("manifest identity drifted")
	}
	if m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		return fmt.Errorf("title, generated_doc, workflow, and evidence_artifact are required")
	}
	return nil
}
