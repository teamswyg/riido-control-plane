package main

import "fmt"

func verifyLoopCoverageComplete(
	loops map[string]map[string]bool,
	covered map[string]map[string]bool,
	dim loopCoverageDimension,
) error {
	for loopID, tokens := range loops {
		for token := range tokens {
			if !covered[loopID][token] {
				return fmt.Errorf("loop %s %s %s is not covered by a claim",
					loopID, dim.loopTokenLabel, token)
			}
		}
	}
	return nil
}
