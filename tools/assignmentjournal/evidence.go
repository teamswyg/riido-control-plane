package main

import "github.com/teamswyg/riido-control-plane/tools/assignmentjournal/requirements"

type evidence struct {
	SchemaVersion string       `json:"schema_version"`
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	Ports         int          `json:"ports"`
	Records       int          `json:"records"`
	ReplayRules   int          `json:"replay_rules"`
	Constants     int          `json:"version_constants"`
	SourceChecks  int          `json:"source_checks"`
	Loop          evidenceLoop `json:"loop"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion: requirements.EvidenceSchema,
		ID:            m.ID,
		Status:        "verified",
		Ports:         len(m.Ports),
		Records:       len(m.Records),
		ReplayRules:   len(m.ReplayRules),
		Constants:     len(m.VersionConstants),
		SourceChecks:  len(m.SourceChecks),
		Loop:          m.Loop,
	}
}
