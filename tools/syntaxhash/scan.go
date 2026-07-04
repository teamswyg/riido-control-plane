package main

import (
	"fmt"
)

func buildGraph(root string, m manifest) (syntaxGraph, error) {
	semantic, err := semanticHashes(root, m.LoopRegistry)
	if err != nil {
		return syntaxGraph{}, err
	}
	graph := syntaxGraph{SchemaVersion: m.SchemaVersion, ID: m.ID, Status: "verified"}
	for _, target := range m.Targets {
		tg, err := scanTarget(root, target, semantic)
		if err != nil {
			return graph, err
		}
		graph.Targets = append(graph.Targets, tg)
	}
	repo, err := scanRepository(root, graph, m.Constraints)
	if err != nil {
		return graph, err
	}
	graph.Repository = repo
	constraints, err := checkConstraints(root, m.ArtifactRoots, m.Constraints)
	if err != nil {
		return graph, err
	}
	graph.Constraints = constraints
	graph.Score = scoreGraph(graph, m)
	return graph, nil
}

func scanTarget(root string, target targetConfig, semantic map[string]string) (targetGraph, error) {
	hash, err := semanticHashFor(semantic, target.SemanticClaimID)
	if err != nil {
		return targetGraph{}, err
	}
	files, err := goFiles(repoPath(root, target.PackagePath))
	if err != nil {
		return targetGraph{}, err
	}
	out := targetGraph{ID: target.ID, PackagePath: target.PackagePath, SemanticClaim: target.SemanticClaimID, SemanticHash: hash}
	seen := map[string]string{}
	for _, file := range files {
		fh, err := scanFile(root, file)
		if err != nil {
			return out, err
		}
		if prev, ok := seen[fh.Hash]; ok && prev != fh.Normalized {
			return out, fmt.Errorf("syntax hash collision for %s", fh.Hash)
		}
		seen[fh.Hash] = fh.Normalized
		out.FileHashes = append(out.FileHashes, fh)
		out.Relocations = append(out.Relocations, relocate{OldHash: fh.Hash, NewLocation: fh.Path})
	}
	out.GoFiles, out.TrackedFiles = len(files), len(out.FileHashes)
	out.Coverage = percent(out.TrackedFiles, out.GoFiles)
	out.PackageHash = packageHash(out.FileHashes)
	return out, nil
}
