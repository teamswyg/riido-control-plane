package main

type chainEvidence struct {
	ID          string   `json:"id"`
	Observation string   `json:"observation"`
	Hypothesis  string   `json:"hypothesis"`
	Claims      []string `json:"claims,omitempty"`
	Changes     []ref    `json:"changes"`
	Verifiers   []ref    `json:"verifiers"`
	Evidence    []ref    `json:"evidence"`
	Decision    string   `json:"decision"`
	NextLoop    string   `json:"next_loop"`
}

func evidenceChains(chains []chain) []chainEvidence {
	out := make([]chainEvidence, 0, len(chains))
	for _, c := range chains {
		out = append(out, chainEvidence(c))
	}
	return out
}
