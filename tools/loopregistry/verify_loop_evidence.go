package main

import "fmt"

func verifyLoopEvidenceCoverage(m manifest) error {
	loops := loopEvidenceTokens(m.Loops)
	covered := map[string]map[string]bool{}
	for _, claim := range m.Claims {
		tokens, ok := loops[claim.Loop]
		if !ok {
			return fmt.Errorf("claim %s binds unknown loop %s", claim.ID, claim.Loop)
		}
		if err := verifyClaimEvidenceCoverage(claim, tokens); err != nil {
			return err
		}
		for _, token := range claim.CoversEvidence {
			if covered[claim.Loop] == nil {
				covered[claim.Loop] = map[string]bool{}
			}
			covered[claim.Loop][token] = true
		}
	}
	for loopID, tokens := range loops {
		for token := range tokens {
			if !covered[loopID][token] {
				return fmt.Errorf("loop %s evidence source %s is not covered by a claim", loopID, token)
			}
		}
	}
	return nil
}

func verifyClaimEvidenceCoverage(claim claimBinding, loopTokens map[string]bool) error {
	for _, token := range claim.CoversEvidence {
		if !loopTokens[token] {
			return fmt.Errorf("claim %s covers unknown evidence source %s", claim.ID, token)
		}
	}
	return nil
}

func loopEvidenceTokens(loops []loopRecord) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, loop := range loops {
		out[loop.ID] = map[string]bool{}
		for _, source := range loop.Evidence {
			out[loop.ID][source.Path] = true
		}
	}
	return out
}
