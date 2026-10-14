package main

import "github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/requirements"

type evidence struct {
	SchemaVersion string          `json:"schema_version"`
	ID            string          `json:"id"`
	Status        string          `json:"status"`
	Counts        operationCounts `json:"operation_counts"`
	SourceChecks  int             `json:"source_checks"`
	EvidenceTests int             `json:"evidence_tests"`
	MatrixParity  bool            `json:"matrix_parity"`
	MatrixSorted  bool            `json:"matrix_sorted"`
	Loop          evidenceLoop    `json:"loop"`
}

func newEvidence(m manifest, counts operationCounts) evidence {
	return evidence{
		SchemaVersion: requirements.EvidenceSchema,
		ID:            m.ID,
		Status:        "verified",
		Counts:        counts,
		SourceChecks:  len(m.SourceChecks),
		EvidenceTests: len(m.RequiredEvidenceTests),
		MatrixParity:  true,
		MatrixSorted:  true,
		Loop:          m.Loop,
	}
}
