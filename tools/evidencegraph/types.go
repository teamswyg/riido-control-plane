package main

type manifest struct {
	SchemaVersion string     `json:"schema_version"`
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	GeneratedDoc  string     `json:"generated_doc"`
	Workflow      string     `json:"workflow"`
	Evidence      string     `json:"evidence_artifact"`
	EvidenceTool  string     `json:"evidence_tool"`
	LoopRegistry  string     `json:"loop_registry_manifest"`
	Assertions    []string   `json:"assertions"`
	Chains        []chain    `json:"chains"`
	Loop          loopRecord `json:"loop"`
}

type chain struct {
	ID          string `json:"id"`
	Observation string `json:"observation"`
	Hypothesis  string `json:"hypothesis"`
	Changes     []ref  `json:"changes"`
	Verifiers   []ref  `json:"verifiers"`
	Evidence    []ref  `json:"evidence"`
	Decision    string `json:"decision"`
	NextLoop    string `json:"next_loop"`
}
