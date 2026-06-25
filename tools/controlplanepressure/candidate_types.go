package main

type candidateEntry struct {
	ID                    string         `json:"id"`
	HarnessLoop           string         `json:"harness_loop"`
	PromotionTarget       string         `json:"promotion_target"`
	Scenario              string         `json:"scenario"`
	Risk                  string         `json:"risk"`
	Next                  string         `json:"next"`
	RequiredNextArtifacts []string       `json:"required_next_artifacts"`
	AdoptionPlan          []adoptionStep `json:"adoption_plan"`
}

type adoptionStep struct {
	Artifact string `json:"artifact"`
	Command  string `json:"command"`
}
