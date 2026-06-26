package main

import (
	"encoding/json"
	"os"
)

func writePressureCandidateEvidence(path string, report pressureReport) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(pressureCandidateEvidenceFromReport(report), "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}
