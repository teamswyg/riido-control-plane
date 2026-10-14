package main

type generator struct {
	ReactQuery string   `json:"react_query"`
	Handoff    string   `json:"handoff"`
	MustNotOwn []string `json:"must_not_own"`
	Artifacts  []string `json:"artifacts"`
}

type figmaContext struct {
	ID             string   `json:"id"`
	NodeIDs        []string `json:"node_ids"`
	GeneratedPaths []string `json:"generated_paths"`
	Rule           string   `json:"rule"`
	MustNotOwn     string   `json:"must_not_own"`
}

type modelCatalog struct {
	Policy      string `json:"policy"`
	Rendering   string `json:"rendering"`
	FixtureRule string `json:"fixture_rule"`
}

type sourceCheck struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

type loopRecord struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
