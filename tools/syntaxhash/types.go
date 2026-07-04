package main

type manifest struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	GeneratedDoc     string         `json:"generated_doc"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	EvidenceTool     string         `json:"evidence_tool"`
	LoopRegistry     string         `json:"loop_registry_manifest"`
	ArtifactRoots    []string       `json:"artifact_roots"`
	Targets          []targetConfig `json:"targets"`
	Constraints      constraints    `json:"constraints"`
	Scoring          scoring        `json:"scoring"`
	Loop             evidenceLoop   `json:"loop"`
}

type targetConfig struct {
	ID              string `json:"id"`
	PackagePath     string `json:"package_path"`
	GoldenCommand   string `json:"golden_command"`
	SemanticClaimID string `json:"semantic_claim_id"`
	Mode            string `json:"mode"`
}

type constraints struct {
	MaxFileLines                    int `json:"max_file_lines"`
	MaxFilesPerFolder               int `json:"max_files_per_folder"`
	MaxDirectoryDepth               int `json:"max_directory_depth"`
	MinCoveragePercent              int `json:"min_coverage_percent"`
	MinRepositoryCoverageBasisPoint int `json:"min_repository_coverage_basis_points"`
	UntrackedSampleLimit            int `json:"untracked_sample_limit"`
}

type scoring struct {
	EfficiencyWeight  int `json:"efficiency_weight"`
	CompressionWeight int `json:"compression_weight"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

func scoreGraph(graph syntaxGraph, cfg scoring) scoreRun {
	files, moves, collisions, hashes := 0, 0, 0, map[string]struct{}{}
	for _, target := range graph.Targets {
		files += target.TrackedFiles
		moves += len(target.Relocations)
		collisions += target.CollisionCount
		for _, file := range target.FileHashes {
			hashes[file.Hash] = struct{}{}
		}
	}
	compression := files - len(hashes)
	efficiency := files * cfg.EfficiencyWeight
	compressed := compression * cfg.CompressionWeight
	return scoreRun{
		TrackedFiles: files, UniqueSyntaxHashes: len(hashes), CompressionGain: compression,
		EfficiencyWeight: cfg.EfficiencyWeight, CompressionWeight: cfg.CompressionWeight,
		EfficiencyScore: efficiency, CompressionScore: compressed, WeightedScore: efficiency + compressed,
		CollisionCount: collisions, RelocationMappings: moves, MissingRelocations: files - moves,
		RelocationCoverage: basisPoints(moves, files),
		ConstraintGate:     "coverage>=floor && collisions==0 && relocations==tracked && physical_violations==0",
		Formula:            "tracked_files*efficiency_weight + compression_gain*compression_weight",
	}
}
