package main

import "fmt"

type verifyResult struct {
	PackageCount        int `json:"package_count"`
	RuntimePackages     int `json:"runtime_packages"`
	InternalPackages    int `json:"internal_packages"`
	ToolPackages        int `json:"tool_packages"`
	ForbiddenImportHits int `json:"forbidden_import_hits"`
	LineBudget          lineBudgetResult
}

func verifyAll(repoRoot string, m manifest) (verifyResult, error) {
	if err := verifyManifestShape(m); err != nil {
		return verifyResult{}, err
	}
	actual, err := scanPackageDirs(repoRoot, m.SourceRoots)
	if err != nil {
		return verifyResult{}, err
	}
	if err := verifyPackageParity(actual, m.Packages); err != nil {
		return verifyResult{}, err
	}
	violations, err := scanForbiddenImports(repoRoot, m.Packages, m.ForbiddenImports)
	if err != nil {
		return verifyResult{}, err
	}
	if len(violations) > 0 {
		return verifyResult{}, fmt.Errorf("forbidden imports: %v", violations)
	}
	result := summarizePackages(m.Packages)
	lineBudget, err := scanLineBudget(repoRoot, m.SourceRoots, m.FileLineBudget)
	if err != nil {
		return verifyResult{}, err
	}
	result.LineBudget = lineBudget
	return result, nil
}

func verifyManifestShape(m manifest) error {
	if m.SchemaVersion == "" || m.ID == "" || m.Title == "" || m.GeneratedDoc == "" {
		return fmt.Errorf("schema_version, id, title, and generated_doc are required")
	}
	if m.ModulePath == "" || len(m.SourceRoots) == 0 || len(m.Packages) == 0 {
		return fmt.Errorf("module_path, source_roots, and packages are required")
	}
	if m.FileLineBudget.TargetLines < 0 || m.FileLineBudget.SampleLimit < 0 {
		return fmt.Errorf("file_line_budget values must be non-negative")
	}
	return verifyLoop(m.Loop)
}
