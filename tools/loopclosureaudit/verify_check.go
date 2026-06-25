package main

import "fmt"

func verifyCheck(root string, c check, idx indexes) error {
	switch c.Kind {
	case "loop":
		return verifyLoopCheck(c, idx)
	case "claim":
		return verifyClaimCheck(c, idx)
	case "workflow":
		return verifyWorkflowCheck(root, c)
	case "graph_edge":
		return verifyGraphEdgeCheck(c, idx)
	case "graph_chain":
		return verifyGraphChainCheck(c, idx)
	case "pre_commit_hook":
		return verifyPreCommitHookCheck(c, idx)
	default:
		return fmt.Errorf("unknown check kind %q", c.Kind)
	}
}

func verifyLoopCheck(c check, idx indexes) error {
	loop, ok := idx.loops[c.ID]
	if !ok {
		return fmt.Errorf("missing loop %s", c.ID)
	}
	if c.MustExpire && (loop.ExpiresAfterHours <= 0 || loop.RefreshWorkflow == "") {
		return fmt.Errorf("loop %s must expire and refresh", c.ID)
	}
	if c.MustPromoteTo != "" && !contains(loop.PromotesTo, c.MustPromoteTo) {
		return fmt.Errorf("loop %s must promote to %s", c.ID, c.MustPromoteTo)
	}
	for _, provider := range c.Providers {
		if !contains(loop.Providers, provider) {
			return fmt.Errorf("loop %s missing provider %s", c.ID, provider)
		}
	}
	return nil
}

func verifyClaimCheck(c check, idx indexes) error {
	claim, ok := idx.claims[c.ID]
	if !ok {
		return fmt.Errorf("missing claim %s", c.ID)
	}
	if claim.SemanticHash == "" || len(claim.Files) == 0 || len(claim.Verifiers) == 0 {
		return fmt.Errorf("claim %s must bind hash, files, and verifiers", c.ID)
	}
	return nil
}
