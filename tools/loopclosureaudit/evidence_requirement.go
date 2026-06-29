package main

func checkCount(requirements []requirement) int {
	count := 0
	for _, req := range requirements {
		count += len(req.Checks)
	}
	return count
}

func requirementEvidenceRows(requirements []requirement, idxOpt ...indexes) []requirementEvidence {
	rows := make([]requirementEvidence, 0, len(requirements))
	for _, req := range requirements {
		proofs := requirementProofs(req.Checks, idxOpt...)
		rows = append(rows, requirementEvidence{
			ID:         req.ID,
			Statement:  req.Statement,
			Status:     "verified",
			CheckKinds: evidenceCheckKinds(req.Checks),
			ProofCount: len(proofs),
			Proofs:     proofs,
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
