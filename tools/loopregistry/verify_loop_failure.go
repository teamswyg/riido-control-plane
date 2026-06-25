package main

import "fmt"

func verifyLoopFailureCoverage(m manifest) error {
	loops := loopFailureTokens(m.Loops)
	covered := map[string]map[string]bool{}
	for _, claim := range m.Claims {
		tokens, ok := loops[claim.Loop]
		if !ok {
			return fmt.Errorf("claim %s binds unknown loop %s", claim.ID, claim.Loop)
		}
		if err := verifyClaimFailureCoverage(claim, tokens); err != nil {
			return err
		}
		for _, token := range claim.CoversFails {
			if covered[claim.Loop] == nil {
				covered[claim.Loop] = map[string]bool{}
			}
			covered[claim.Loop][token] = true
		}
	}
	for loopID, tokens := range loops {
		for token := range tokens {
			if !covered[loopID][token] {
				return fmt.Errorf("loop %s fail token %s is not covered by a claim", loopID, token)
			}
		}
	}
	return nil
}

func verifyClaimFailureCoverage(claim claimBinding, loopTokens map[string]bool) error {
	for _, token := range claim.CoversFails {
		if !loopTokens[token] {
			return fmt.Errorf("claim %s covers unknown fail token %s", claim.ID, token)
		}
	}
	return nil
}

func loopFailureTokens(loops []loopRecord) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, loop := range loops {
		out[loop.ID] = map[string]bool{}
		for _, token := range loop.FailsWhen {
			out[loop.ID][token] = true
		}
	}
	return out
}
