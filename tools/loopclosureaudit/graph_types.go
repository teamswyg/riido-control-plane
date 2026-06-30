package main

type evidenceGraph struct {
	Chains []evidenceChain `json:"chains"`
}

type evidenceChain struct {
	ID          string   `json:"id"`
	Observation string   `json:"observation"`
	Hypothesis  string   `json:"hypothesis"`
	Claims      []string `json:"claims"`
	Changes     []ref    `json:"changes"`
	Verifiers   []ref    `json:"verifiers"`
	Evidence    []ref    `json:"evidence"`
	Decision    string   `json:"decision"`
	NextLoop    string   `json:"next_loop"`
}

type ref struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}
