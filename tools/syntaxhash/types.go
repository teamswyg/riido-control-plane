package main

type manifest struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	EvidenceTool     string          `json:"evidence_tool"`
	LoopRegistry     string          `json:"loop_registry_manifest"`
	ArtifactRoots    []string        `json:"artifact_roots"`
	DuplicateShapes  duplicatePolicy `json:"duplicate_shape_policy"`
	Targets          []targetConfig  `json:"targets"`
	Constraints      constraints     `json:"constraints"`
	Scoring          struct {
		EfficiencyWeight  int `json:"efficiency_weight"`
		CompressionWeight int `json:"compression_weight"`
	} `json:"scoring"`
	Loop evidenceLoop `json:"loop"`
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

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

func scoreGraph(graph syntaxGraph, m manifest) scoreRun {
	return scoreGraphRun(graph, m)
}

type scoreRun struct {
	TrackedFiles       int    `json:"tracked_files"`
	UniqueSyntaxHashes int    `json:"unique_syntax_hashes"`
	CompressionGain    int    `json:"compression_gain"`
	AnalysisReduction  int    `json:"analysis_reduction_basis_points"`
	EfficiencyWeight   int    `json:"efficiency_weight"`
	CompressionWeight  int    `json:"compression_weight"`
	EfficiencyScore    int    `json:"efficiency_score"`
	CompressionScore   int    `json:"compression_score"`
	WeightedScore      int    `json:"weighted_score"`
	CollisionCount     int    `json:"collision_count"`
	RelocationMappings int    `json:"relocation_mappings"`
	MissingRelocations int    `json:"missing_relocation_mappings"`
	RelocationCoverage int    `json:"relocation_coverage_basis_points"`
	GoldenCommands     int    `json:"golden_commands"`
	MissingGoldens     int    `json:"missing_golden_commands"`
	ConstraintGate     string `json:"constraint_gate"`
	Formula            string `json:"formula"`
}
