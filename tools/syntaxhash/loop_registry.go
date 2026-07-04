package main

import "fmt"

type loopRegistryClaim struct {
	ID           string `json:"id"`
	SemanticHash string `json:"semantic_hash"`
}

type loopRegistryManifest struct {
	Claims []loopRegistryClaim `json:"claim_bindings"`
}

func semanticHashes(root, path string) (map[string]string, error) {
	var m loopRegistryManifest
	if err := readJSON(repoPath(root, path), &m); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, claim := range m.Claims {
		if claim.ID != "" && claim.SemanticHash != "" {
			out[claim.ID] = claim.SemanticHash
		}
	}
	return out, nil
}

func semanticHashFor(values map[string]string, claimID string) (string, error) {
	value := values[claimID]
	if value == "" {
		return "", fmt.Errorf("semantic claim %s has no hash", claimID)
	}
	return value, nil
}

func scoreGraphRun(graph syntaxGraph, m manifest) scoreRun {
	files, moves, collisions, goldens, hashes := 0, 0, 0, 0, map[string]struct{}{}
	for _, target := range graph.Targets {
		files += target.TrackedFiles
		moves += len(target.Relocations)
		collisions += target.CollisionCount
		for _, file := range target.FileHashes {
			hashes[file.Hash] = struct{}{}
		}
	}
	for _, target := range m.Targets {
		if target.GoldenCommand != "" {
			goldens++
		}
	}
	return scoreGraphResult(files, moves, collisions, goldens, len(hashes), m)
}

func scoreGraphResult(files, moves, collisions, goldens, unique int, m manifest) scoreRun {
	compression := files - unique
	reduction := basisPoints(compression, files)
	return scoreRun{
		TrackedFiles: files, UniqueSyntaxHashes: unique, CompressionGain: compression,
		AnalysisReduction: reduction,
		EfficiencyWeight:  m.Scoring.EfficiencyWeight, CompressionWeight: m.Scoring.CompressionWeight,
		EfficiencyScore:  reduction * m.Scoring.EfficiencyWeight,
		CompressionScore: compression * m.Scoring.CompressionWeight,
		WeightedScore:    reduction*m.Scoring.EfficiencyWeight + compression*m.Scoring.CompressionWeight,
		CollisionCount:   collisions, RelocationMappings: moves, MissingRelocations: files - moves,
		RelocationCoverage: basisPoints(moves, files),
		GoldenCommands:     goldens, MissingGoldens: len(m.Targets) - goldens,
		ConstraintGate: "coverage>=floor && golden_commands==targets && collisions==0 && relocations==tracked && physical_violations==0",
		Formula:        "analysis_reduction_basis_points*efficiency_weight + compression_gain*compression_weight",
	}
}
