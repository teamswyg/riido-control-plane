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
		len(document.BoundedChildren) != 6 {
		return manifest{}, errors.New("baseline CI parity contract identity drifted")
	}
	child := document.BoundedChildren[0]
	if child.ID != "control-plane-repository-readme-ci-parity" || child.Issue != readmeIssueURL ||
		child.ParentIssue != parentIssueURL || len(child.Assertions) < 5 || !completeLoop(child.Loop) {
		return manifest{}, errors.New("repository README parity child identity drifted")
	}
	contextChild := document.BoundedChildren[1]
	if contextChild.ID != "control-plane-context-map-ci-parity" || contextChild.Issue != contextIssueURL ||
		contextChild.ParentIssue != parentIssueURL || len(contextChild.Assertions) < 5 ||
		!completeLoop(contextChild.Loop) {
		return manifest{}, errors.New("context map parity child identity drifted")
	}
	goCIChild := document.BoundedChildren[2]
	if goCIChild.ID != "control-plane-go-ci-quality-parity" || goCIChild.Issue != goCIIssueURL ||
		goCIChild.ParentIssue != parentIssueURL || len(goCIChild.Assertions) < 5 ||
		!completeLoop(goCIChild.Loop) {
		return manifest{}, errors.New("go CI quality parity child identity drifted")
	}
	moduleChild := document.BoundedChildren[3]
	if moduleChild.ID != "control-plane-module-decomposition-ci-parity" || moduleChild.Issue != moduleIssueURL ||
		moduleChild.ParentIssue != parentIssueURL || len(moduleChild.Assertions) < 5 ||
		!completeLoop(moduleChild.Loop) {
		return manifest{}, errors.New("module decomposition parity child identity drifted")
	}
	preCommitChild := document.BoundedChildren[4]
	if preCommitChild.ID != "control-plane-pre-commit-baseline-ci-parity" ||
		preCommitChild.Issue != preCommitIssueURL || preCommitChild.ParentIssue != parentIssueURL ||
		len(preCommitChild.Assertions) < 5 || !completeLoop(preCommitChild.Loop) {
		return manifest{}, errors.New("pre-commit baseline parity child identity drifted")
	}
	migrationChild := document.BoundedChildren[5]
	if migrationChild.ID != "control-plane-migration-ledger-ci-parity" ||
		migrationChild.Issue != migrationIssueURL || migrationChild.ParentIssue != parentIssueURL ||
		len(migrationChild.Assertions) < 5 || !completeLoop(migrationChild.Loop) {
		return manifest{}, errors.New("migration ledger parity child identity drifted")
	}
	return document, nil
}

func completeLoop(loop evidenceLoop) bool {
	return loop.Observation != "" && loop.Hypothesis != "" && loop.Execute != "" &&
		loop.Evaluate != "" && loop.Retrospective != ""
}
