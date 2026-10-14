package main

type manifest struct {
	SchemaVersion         string            `json:"schema_version"`
	ID                    string            `json:"id"`
	Title                 string            `json:"title"`
	GeneratedDoc          string            `json:"generated_doc"`
	Workflow              string            `json:"workflow"`
	EvidenceArtifact      string            `json:"evidence_artifact"`
	EvidenceTool          string            `json:"evidence_tool"`
	LoopRegistryManifest  string            `json:"loop_registry_manifest"`
	EvidenceGraphManifest string            `json:"evidence_graph_manifest"`
	PreCommitManifest     string            `json:"pre_commit_manifest"`
	Sources               []candidateSource `json:"sources"`
	Requirements          []requirement     `json:"requirements"`
	Assertions            []string          `json:"assertions"`
	ResidualGaps          []residualGap     `json:"residual_gaps"`
	Loop                  loopSpec          `json:"loop"`
}

type requirement struct {
	ID        string  `json:"id"`
	Statement string  `json:"statement"`
	Checks    []check `json:"checks"`
}

type check struct {
	Kind          string   `json:"kind"`
	ID            string   `json:"id,omitempty"`
	Path          string   `json:"path,omitempty"`
	From          string   `json:"from,omitempty"`
	To            string   `json:"to,omitempty"`
	Relation      string   `json:"relation,omitempty"`
	MustPromoteTo string   `json:"must_promote_to,omitempty"`
	Contains      []string `json:"contains,omitempty"`
	Providers     []string `json:"providers,omitempty"`
	Claims        []string `json:"claims,omitempty"`
	MustExpire    bool     `json:"must_expire,omitempty"`
}

type residualGap struct {
	ID           string `json:"id"`
	Observation  string `json:"observation"`
	Risk         string `json:"risk"`
	NextLoop     string `json:"next_loop"`
	NextArtifact string `json:"next_artifact"`
	NextCommand  string `json:"next_command"`
}

type candidateSource struct {
	ID                    string   `json:"id"`
	SourceWorkflow        string   `json:"source_workflow"`
	SummaryArtifact       string   `json:"summary_artifact"`
	CandidateArtifact     string   `json:"candidate_artifact"`
	HarnessLoop           string   `json:"harness_loop"`
	PromotionTarget       string   `json:"promotion_target"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
}

type loopSpec struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
