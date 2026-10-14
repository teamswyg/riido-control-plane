package main

type syntaxGraph struct {
	SchemaVersion string        `json:"schema_version"`
	ID            string        `json:"id"`
	Status        string        `json:"status"`
	Repository    repoCoverage  `json:"repository_coverage"`
	Targets       []targetGraph `json:"targets"`
	Duplicates    duplicateRun  `json:"duplicate_shapes"`
	Constraints   constraintRun `json:"constraints"`
	Score         scoreRun      `json:"score"`
}

type repoCoverage struct {
	GoFiles             int      `json:"go_files"`
	TrackedFiles        int      `json:"tracked_files"`
	UntrackedFiles      int      `json:"untracked_files"`
	CoveragePercent     int      `json:"coverage_percent"`
	CoverageBasisPoints int      `json:"coverage_basis_points"`
	UntrackedSample     []string `json:"untracked_sample"`
}

type targetGraph struct {
	ID             string     `json:"id"`
	PackagePath    string     `json:"package_path"`
	GoFiles        int        `json:"go_files"`
	TrackedFiles   int        `json:"tracked_files"`
	Coverage       int        `json:"coverage_percent"`
	PackageHash    string     `json:"package_syntax_hash"`
	SemanticClaim  string     `json:"semantic_claim_id"`
	SemanticHash   string     `json:"semantic_hash"`
	FileHashes     []fileHash `json:"file_hashes"`
	CollisionCount int        `json:"collision_count"`
	Relocations    []relocate `json:"relocations"`
}

type fileHash struct {
	Path       string `json:"path"`
	Hash       string `json:"syntax_hash"`
	ShapeHash  string `json:"shape_hash"`
	Shape      string `json:"shape"`
	Normalized string `json:"normalized"`
}

type relocate struct {
	OldHash     string `json:"old_hash"`
	NewLocation string `json:"new_location"`
}

type constraintRun struct {
	MaxFileLines                    int `json:"max_file_lines"`
	MaxFilesPerFolder               int `json:"max_files_per_folder"`
	MaxDepth                        int `json:"max_directory_depth"`
	MinRepositoryCoverageBasisPoint int `json:"min_repository_coverage_basis_points"`
	Violations                      int `json:"violations"`
}
