package main

func checkCount(requirements []requirement) int {
	count := 0
	for _, req := range requirements {
		count += len(req.Checks)
	}
	return count
}

func requirementEvidenceRows(requirements []requirement) []requirementEvidence {
	rows := make([]requirementEvidence, 0, len(requirements))
	for _, req := range requirements {
		rows = append(rows, requirementEvidence{
			ID:         req.ID,
			Statement:  req.Statement,
			CheckKinds: evidenceCheckKinds(req.Checks),
			Checks:     evidenceChecks(req.Checks),
		})
	}
	return rows
}

func evidenceCheckKinds(checks []check) []string {
	kinds := make([]string, 0, len(checks))
	for _, check := range checks {
		kinds = append(kinds, check.Kind)
	}
	return kinds
}
