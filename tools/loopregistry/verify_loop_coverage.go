package main

import "fmt"

func verifyLoopVerifyCoverage(m manifest) error {
	loops := loopVerifyTokens(m.Loops)
	covered := map[string]map[string]bool{}
	for _, claim := range m.Claims {
		tokens, ok := loops[claim.Loop]
		if !ok {
			return fmt.Errorf("claim %s binds unknown loop %s", claim.ID, claim.Loop)
		}
		if err := verifyClaimVerifyCoverage(claim, tokens); err != nil {
			return err
		}
		for _, token := range claim.CoversVerifies {
			if covered[claim.Loop] == nil {
				covered[claim.Loop] = map[string]bool{}
			}
			covered[claim.Loop][token] = true
		}
	}
	for loopID, tokens := range loops {
		for token := range tokens {
			if !covered[loopID][token] {
				return fmt.Errorf("loop %s verify token %s is not covered by a claim", loopID, token)
			}
		}
	}
	return nil
}

func verifyClaimVerifyCoverage(claim claimBinding, loopTokens map[string]bool) error {
	if len(claim.CoversVerifies) == 0 {
		return nil
	}
	for _, token := range claim.CoversVerifies {
		if !loopTokens[token] {
			return fmt.Errorf("claim %s covers unknown verify token %s", claim.ID, token)
		}
	}
	return nil
}

func loopVerifyTokens(loops []loopRecord) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, loop := range loops {
		out[loop.ID] = map[string]bool{}
		for _, token := range loop.Verifies {
			out[loop.ID][token] = true
		}
	}
	return out
}
