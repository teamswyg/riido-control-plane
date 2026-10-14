package main

import "fmt"

func verifyLoopCoverageDimension(m manifest, dim loopCoverageDimension) error {
	loops := loopCoverageTokens(m.Loops, dim)
	covered := map[string]map[string]bool{}
	for _, claim := range m.Claims {
		tokens, ok := loops[claim.Loop]
		if !ok {
			return fmt.Errorf("claim %s binds unknown loop %s", claim.ID, claim.Loop)
		}
		if err := verifyClaimCoverageDimension(claim, dim, tokens); err != nil {
			return err
		}
		for _, token := range dim.claimTokens(claim) {
			if covered[claim.Loop] == nil {
				covered[claim.Loop] = map[string]bool{}
			}
			covered[claim.Loop][token] = true
		}
	}
	return verifyLoopCoverageComplete(loops, covered, dim)
}

func verifyClaimCoverageDimension(
	claim claimBinding,
	dim loopCoverageDimension,
	loopTokens map[string]bool,
) error {
	for _, token := range dim.claimTokens(claim) {
		if !loopTokens[token] {
			return fmt.Errorf("claim %s covers unknown %s %s",
				claim.ID, dim.claimTokenLabel, token)
		}
	}
	return nil
}
