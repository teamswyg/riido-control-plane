package main

type evidence struct {
	SchemaVersion        string          `json:"schema_version"`
	ID                   string          `json:"id"`
	Status               string          `json:"status"`
	OperationCounts      operationCounts `json:"operation_counts"`
	RequiredPaths        int             `json:"required_generated_paths"`
	RuntimeConfigs       int             `json:"runtime_configs"`
	PublicFields         int             `json:"public_fields"`
	DeploymentEvidence   int             `json:"deployment_evidence"`
	ThreadHistoryV3Rules int             `json:"thread_history_v3_rules"`
	SourceChecks         int             `json:"source_checks"`
	SmokeMatrixParity    bool            `json:"smoke_matrix_parity"`
	GeneratedPathCovered bool            `json:"generated_path_covered"`
	Loop                 loop            `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion:        evidenceSchema,
		ID:                   m.ID,
		Status:               "verified",
		OperationCounts:      m.OperationCounts,
		RequiredPaths:        len(m.RequiredGeneratedPaths),
		RuntimeConfigs:       len(m.RuntimeConfigKeys),
		PublicFields:         len(m.PublicFields),
		DeploymentEvidence:   len(m.DeploymentEvidence),
		ThreadHistoryV3Rules: len(m.ThreadHistoryV3.Checklist),
		SourceChecks:         len(m.SourceChecks),
		SmokeMatrixParity:    true,
		GeneratedPathCovered: true,
		Loop:                 m.Loop,
	}
}
