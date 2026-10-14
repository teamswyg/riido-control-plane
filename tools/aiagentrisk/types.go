package main

type evidenceManifest struct {
	SchemaVersion     string              `json:"schema_version"`
	ID                string              `json:"id"`
	RiidoTask         string              `json:"riido_task"`
	HumanDoc          string              `json:"human_doc"`
	LocalEvidence     []localEvidence     `json:"local_evidence"`
	ExternalEvidence  []externalEvidence  `json:"external_evidence"`
	RemainingBoundary []remainingBoundary `json:"remaining_boundaries"`
	Loop              evidenceLoop        `json:"loop"`
}

type localEvidence struct {
	Risk    string `json:"risk"`
	Status  string `json:"status"`
	Package string `json:"package"`
	Test    string `json:"test"`
	Proves  string `json:"proves"`
}

type externalEvidence struct {
	Risk   string `json:"risk"`
	Status string `json:"status"`
	Repo   string `json:"repo"`
	Test   string `json:"test"`
	Proves string `json:"proves"`
}

type remainingBoundary struct {
	ID     string `json:"id"`
	Owner  string `json:"owner"`
	Reason string `json:"reason"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
