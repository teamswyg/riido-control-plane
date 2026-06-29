package main

func proofCount(requirements []requirementEvidence) int {
	count := 0
	for _, req := range requirements {
		count += req.ProofCount
	}
	return count
}

func proofSurfaceCount(requirements []requirementEvidence) int {
	count := 0
	for _, req := range requirements {
		count += req.ProofSurfaceCount
	}
	return count
}

func proofSurfaceGapCount(requirements []requirementEvidence) int {
	return proofCount(requirements) - proofSurfaceCount(requirements)
}

func proofSurfaceCountFor(proofs []proof) int {
	count := 0
	for _, proof := range proofs {
		if proof.Surface != nil {
			count++
		}
	}
	return count
}
