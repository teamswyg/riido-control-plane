package main

func claimCheckProofSurface(c check, idx indexes) *proofSurface {
	return claimProofSurface(c.ID, idx)
}

func loopCheckProofSurface(c check, idx indexes) *proofSurface {
	return loopProofSurface(c.ID, idx)
}

func workflowCheckProofSurface(c check, _ indexes) *proofSurface {
	return &proofSurface{
		Workflow: c.Path,
		Contains: append([]string(nil), c.Contains...),
	}
}

func graphChainCheckProofSurface(c check, idx indexes) *proofSurface {
	return graphChainProofSurface(c.ID, idx)
}

func graphEdgeCheckProofSurface(c check, _ indexes) *proofSurface {
	return graphEdgeProofSurface(c)
}

func graphSummaryCheckProofSurface(_ check, idx indexes) *proofSurface {
	return graphSummaryProofSurface(idx)
}

func harnessSummaryCheckProofSurface(_ check, idx indexes) *proofSurface {
	return harnessSummaryProofSurface(idx)
}

func preCommitHookCheckProofSurface(c check, idx indexes) *proofSurface {
	return preCommitHookProofSurface(c.ID, idx)
}
