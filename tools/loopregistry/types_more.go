package main

type claimBinding struct {
	ID           string   `json:"id"`
	Statement    string   `json:"statement"`
	Loop         string   `json:"loop"`
	Files        []string `json:"files"`
	Verifiers    []string `json:"verifiers"`
	GeneratedDoc []string `json:"generated_docs"`
	SemanticHash string   `json:"semantic_hash"`
}

type graphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type verifyResult struct {
	Loops          int
	Harnesses      int
	ClosedLoops    int
	Claims         int
	GraphEdges     int
	MaxExpiryHours int
	Hashes         map[string]string
}
