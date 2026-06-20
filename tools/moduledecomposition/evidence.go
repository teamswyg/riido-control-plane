package main

type evidence struct {
	SchemaVersion      string       `json:"schema_version"`
	ID                 string       `json:"id"`
	Status             string       `json:"status"`
	PackageCount       int          `json:"package_count"`
	RuntimePackages    int          `json:"runtime_packages"`
	InternalPackages   int          `json:"internal_packages"`
	ToolPackages       int          `json:"tool_packages"`
	ForbiddenImportHit int          `json:"forbidden_import_hits"`
	Workflow           string       `json:"workflow"`
	GeneratedDoc       string       `json:"generated_doc"`
	Loop               evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:      evidenceSchema,
		ID:                 m.ID,
		Status:             "verified",
		PackageCount:       result.PackageCount,
		RuntimePackages:    result.RuntimePackages,
		InternalPackages:   result.InternalPackages,
		ToolPackages:       result.ToolPackages,
		ForbiddenImportHit: result.ForbiddenImportHits,
		Workflow:           m.Workflow,
		GeneratedDoc:       m.GeneratedDoc,
		Loop:               m.Loop,
	}
}
