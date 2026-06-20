package main

type manifest struct {
	SchemaVersion            string              `json:"schema_version"`
	ID                       string              `json:"id"`
	RiidoTask                string              `json:"riido_task"`
	HumanDoc                 string              `json:"human_doc"`
	Decision                 decision            `json:"decision"`
	OperationEvidence        []operationEvidence `json:"operation_evidence"`
	MeasurementSignals       []string            `json:"measurement_signals"`
	CadenceRules             []cadenceRule       `json:"cadence_rules"`
	DecisionRules            []decisionRule      `json:"decision_rules"`
	CandidateSplit           candidateSplit      `json:"candidate_split"`
	ForbiddenTraceAttributes []string            `json:"forbidden_trace_attributes"`
	NonGoals                 []string            `json:"non_goals"`
	Loop                     evidenceLoop        `json:"loop"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

type decision struct {
	Scope         string `json:"scope"`
	StoreWideCQRS bool   `json:"store_wide_cqrs"`
	Reason        string `json:"reason"`
}

type operationEvidence struct {
	Route           string   `json:"route"`
	Intent          string   `json:"intent"`
	StoreOperations []string `json:"store_operations"`
}

type cadenceRule struct {
	Name                 string `json:"name"`
	Seconds              int    `json:"seconds"`
	MustStayBelowSeconds int    `json:"must_stay_below_seconds"`
}

type decisionRule struct {
	ID                   string `json:"id"`
	ThresholdDropPercent int    `json:"threshold_drop_percent"`
	When                 string `json:"when"`
	Action               string `json:"action"`
}

type candidateSplit struct {
	CommandModels []string `json:"command_models"`
	QueryModels   []string `json:"query_models"`
}
