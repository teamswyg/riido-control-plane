package main

func proofSurfaceFor(c check, idx indexes) *proofSurface {
	switch c.Kind {
	case "claim":
		return claimProofSurface(c.ID, idx)
	case "loop":
		return loopProofSurface(c.ID, idx)
	case "workflow":
		return &proofSurface{Workflow: c.Path, Contains: append([]string(nil), c.Contains...)}
	case "graph_chain":
		return graphChainProofSurface(c.ID, idx)
	default:
		return nil
	}
}

func claimProofSurface(id string, idx indexes) *proofSurface {
	claim, ok := idx.claims[id]
	if !ok {
		return nil
	}
	return &proofSurface{
		Files:         append([]string(nil), claim.Files...),
		Verifiers:     append([]string(nil), claim.Verifiers...),
		GeneratedDocs: append([]string(nil), claim.GeneratedDocs...),
		SemanticHash:  claim.SemanticHash,
	}
}

func loopProofSurface(id string, idx indexes) *proofSurface {
	loop, ok := idx.loops[id]
	if !ok {
		return nil
	}
	return &proofSurface{
		Observes:          append([]string(nil), loop.Observes...),
		Verifies:          append([]string(nil), loop.Verifies...),
		FailsWhen:         append([]string(nil), loop.FailsWhen...),
		RefreshWorkflow:   loop.RefreshWorkflow,
		ExpiresAfterHours: loop.ExpiresAfterHours,
		Providers:         append([]string(nil), loop.Providers...),
		PromotesTo:        append([]string(nil), loop.PromotesTo...),
	}
}
