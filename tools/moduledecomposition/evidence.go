package main

type evidence struct {
	SchemaVersion               string                     `json:"schema_version"`
	ID                          string                     `json:"id"`
	Status                      string                     `json:"status"`
	PackageCount                int                        `json:"package_count"`
	RuntimePackages             int                        `json:"runtime_packages"`
	InternalPackages            int                        `json:"internal_packages"`
	ToolPackages                int                        `json:"tool_packages"`
	ForbiddenImportHit          int                        `json:"forbidden_import_hits"`
	LineBudgetTarget            int                        `json:"line_budget_target"`
	FilesOverLineBudget         int                        `json:"files_over_line_budget"`
	MaxFileLines                int                        `json:"max_file_lines"`
	LineBudgetRatchet           lineBudgetRatchet          `json:"line_budget_ratchet"`
	LineBudgetSamples           []lineBudgetSample         `json:"line_budget_samples"`
	LineBudgetHotspots          []lineBudgetHotspot        `json:"line_budget_hotspots"`
	LineBudgetHotspotRatchets   []lineBudgetHotspotRatchet `json:"line_budget_hotspot_ratchets"`
	LineBudgetUntrackedHotspots []lineBudgetHotspot        `json:"line_budget_untracked_hotspots"`
	Workflow                    string                     `json:"workflow"`
	GeneratedDoc                string                     `json:"generated_doc"`
	Loop                        evidenceLoop               `json:"loop"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:               evidenceSchema,
		ID:                          m.ID,
		Status:                      "verified",
		PackageCount:                result.PackageCount,
		RuntimePackages:             result.RuntimePackages,
		InternalPackages:            result.InternalPackages,
		ToolPackages:                result.ToolPackages,
		ForbiddenImportHit:          result.ForbiddenImportHits,
		LineBudgetTarget:            result.LineBudget.Target,
		FilesOverLineBudget:         result.LineBudget.OverTarget,
		MaxFileLines:                result.LineBudget.MaxLines,
		LineBudgetRatchet:           newLineBudgetRatchet(result.LineBudget),
		LineBudgetSamples:           result.LineBudget.Samples,
		LineBudgetHotspots:          result.LineBudget.Hotspots,
		LineBudgetHotspotRatchets:   result.LineBudget.HotspotRatchets,
		LineBudgetUntrackedHotspots: result.LineBudget.UntrackedHotspots,
		Workflow:                    m.Workflow,
		GeneratedDoc:                m.GeneratedDoc,
		Loop:                        m.Loop,
	}
}
