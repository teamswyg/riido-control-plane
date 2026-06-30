package main

type indexes struct {
	loops           map[string]registryLoop
	claims          map[string]registryClaim
	edges           map[graphEdge]struct{}
	chains          map[string]evidenceChain
	hooks           map[string]preCommitHook
	graphSummary    graphChainSummary
	nextLoopSummary []graphNextLoopSummary
	harnessLoops    []registryLoop
}

func newIndexes(deps dependencies) indexes {
	idx := indexes{
		loops:           map[string]registryLoop{},
		claims:          map[string]registryClaim{},
		edges:           map[graphEdge]struct{}{},
		chains:          map[string]evidenceChain{},
		hooks:           map[string]preCommitHook{},
		graphSummary:    summarizeGraphChains(deps.graph.Chains),
		nextLoopSummary: summarizeGraphNextLoops(deps.graph.Chains),
	}
	for _, loop := range deps.registry.Loops {
		idx.loops[loop.ID] = loop
		if loop.Kind == "harness" {
			idx.harnessLoops = append(idx.harnessLoops, loop)
		}
	}
	for _, claim := range deps.registry.Claims {
		idx.claims[claim.ID] = claim
	}
	for _, edge := range deps.registry.Edges {
		idx.edges[edge] = struct{}{}
	}
	for _, chain := range deps.graph.Chains {
		idx.chains[chain.ID] = chain
	}
	for _, hook := range deps.preCommit.Hooks {
		idx.hooks[hook.ID] = hook
	}
	return idx
}
