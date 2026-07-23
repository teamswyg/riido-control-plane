package main

import (
	"errors"
	"fmt"
)

func loadManifest(repoRoot, path string) (manifest, error) {
	var document manifest
	if err := readStrictJSON(repoRoot, path, &document); err != nil {
		return manifest{}, fmt.Errorf("load baseline CI parity contract: %w", err)
	}
	if document.SchemaVersion != contractSchema ||
		document.ID != "control-plane-baseline-go-ci-parity" ||
		document.Status != "parity-candidate-runtime-unchanged" ||
		document.Issue != issueURL || document.ParentIssue != parentIssueURL ||
		document.CheckedOn != "2026-07-22" || len(document.Assertions) < 5 ||
		!completeLoop(document.Loop) ||
		len(document.Pipelines) != 1 || document.Pipelines[0] != document.Runner.Pipeline ||
		len(document.BoundedChildren) != 9 {
		return manifest{}, errors.New("baseline CI parity contract identity drifted")
	}
	if err := verifyBoundedChildIdentities(document); err != nil {
		return manifest{}, err
	}
	return document, nil
}

func completeLoop(loop evidenceLoop) bool {
	return loop.Observation != "" && loop.Hypothesis != "" && loop.Execute != "" &&
		loop.Evaluate != "" && loop.Retrospective != ""
}
