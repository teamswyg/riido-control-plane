package aiagentrisk

type evidenceManifest struct {
	SchemaVersion     string              `json:"schema_version"`
	ID                string              `json:"id"`
	RiidoTask         string              `json:"riido_task"`
	HumanDoc          string              `json:"human_doc"`
	LocalEvidence     []localEvidence     `json:"local_evidence"`
	RemainingBoundary []remainingBoundary `json:"remaining_boundaries"`
}

type localEvidence struct {
	Risk    string `json:"risk"`
	Status  string `json:"status"`
	Package string `json:"package"`
	Test    string `json:"test"`
	Proves  string `json:"proves"`
}

type remainingBoundary struct {
	ID     string `json:"id"`
	Owner  string `json:"owner"`
	Reason string `json:"reason"`
}
