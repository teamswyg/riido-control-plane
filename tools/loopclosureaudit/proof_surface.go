package main

func proofSurfaceFor(c check, idx indexes) *proofSurface {
	spec, ok := checkKindByName(c.Kind)
	if !ok {
		return nil
	}
	return spec.surface(c, idx)
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

func workflowProofSurface(c check) *proofSurface {
	return &proofSurface{
		Workflow: c.Path,
		Contains: append([]string(nil), c.Contains...),
	}
}

func graphEdgeProofSurface(c check) *proofSurface {
	return &proofSurface{From: c.From, To: c.To, Relation: c.Relation}
}

func preCommitHookProofSurface(id string, idx indexes) *proofSurface {
	hook, ok := idx.hooks[id]
	if !ok {
		return nil
	}
	return &proofSurface{
		PreCommitHook: hook.ID,
		Summary:       hook.Summary,
		Contains:      append([]string(nil), hook.Contains...),
	}
}
