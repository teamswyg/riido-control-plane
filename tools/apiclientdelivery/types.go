package main

type manifest struct {
	SchemaVersion string         `json:"schema_version"`
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	RiidoTask     string         `json:"riido_task"`
	GeneratedDoc  string         `json:"generated_doc"`
	Workflow      string         `json:"workflow"`
	Evidence      string         `json:"evidence_artifact"`
	RiskEvidence  string         `json:"risk_evidence_manifest"`
	Sources       []sourceRef    `json:"source_manifests"`
	Owners        []owner        `json:"owners"`
	Delivery      delivery       `json:"delivery_workflow"`
	Branch        branchRule     `json:"target_branch"`
	Lifecycle     []string       `json:"lifecycle_states"`
	Generator     generator      `json:"generator_boundary"`
	Figma         []figmaContext `json:"figma_contexts"`
	ModelCatalog  modelCatalog   `json:"runtime_model_catalog"`
	Required      []string       `json:"required_doc_phrases"`
	Forbidden     []string       `json:"forbidden_doc_phrases"`
	Checks        []sourceCheck  `json:"source_checks"`
	Loop          loopRecord     `json:"loop"`
	NonGoals      []string       `json:"non_goals"`
}

type sourceRef struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type owner struct {
	Name       string `json:"name"`
	Owns       string `json:"owns"`
	DoesNotOwn string `json:"does_not_own"`
}

type delivery struct {
	Workflow           string `json:"workflow"`
	PackageMode        string `json:"package_mode"`
	DeliverMode        string `json:"deliver_mode"`
	IntentionalFailure string `json:"intentional_failure"`
}

type branchRule struct {
	Source     string `json:"source"`
	Rule       string `json:"rule"`
	Example    string `json:"example"`
	SecretGate string `json:"secret_gate"`
}
