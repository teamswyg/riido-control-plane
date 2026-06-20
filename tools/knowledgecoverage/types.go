package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	ScanRoots        []string      `json:"scan_roots"`
	ManualGroups     []manualGroup `json:"manual_groups"`
	Assertions       []string      `json:"assertions"`
	Loop             evidenceLoop  `json:"loop"`
}

type manualGroup struct {
	ID           string   `json:"id"`
	Owner        string   `json:"owner"`
	Reason       string   `json:"reason"`
	NextArtifact string   `json:"next_artifact"`
	Paths        []string `json:"paths,omitempty"`
	PathPrefixes []string `json:"path_prefixes,omitempty"`
}

type docClass struct {
	Path    string
	Kind    string
	Group   string
	Reason  string
	HasLoop bool
}

type evidence struct {
	SchemaVersion     string         `json:"schema_version"`
	ID                string         `json:"id"`
	Status            string         `json:"status"`
	ScannedCount      int            `json:"scanned_count"`
	GeneratedCount    int            `json:"generated_count"`
	DirectSSOTCount   int            `json:"direct_ssot_count"`
	ManualCount       int            `json:"manual_count"`
	ManualByGroup     map[string]int `json:"manual_by_group"`
	ManualTopDirs     []manualDir    `json:"manual_top_dirs"`
	ManualSamples     []manualSample `json:"manual_samples"`
	DirectLoopCount   int            `json:"direct_ssot_loop_count"`
	DirectMissingLoop []string       `json:"direct_ssot_missing_loop"`
	ProblemSummaries  []string       `json:"problem_summaries"`
	EvidenceArtifact  string         `json:"evidence_artifact"`
	Loop              evidenceLoop   `json:"loop"`
}

type manualDir struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type manualSample struct {
	Group  string `json:"group"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
