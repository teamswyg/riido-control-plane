package main

type burstFinding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
}

func hasFinding(findings []burstFinding, id, severity string) bool {
	for _, finding := range findings {
		if finding.ID == id && finding.Severity == severity {
			return true
		}
	}
	return false
}
