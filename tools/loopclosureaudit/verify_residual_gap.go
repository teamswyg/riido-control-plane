package main

import (
	"fmt"
	"strings"
)

func verifyResidualGaps(gaps []residualGap, idx indexes) error {
	if len(gaps) == 0 {
		return fmt.Errorf("audit must declare residual gap candidates")
	}
	seen := map[string]struct{}{}
	for _, gap := range gaps {
		if err := verifyResidualGap(gap, idx); err != nil {
			return err
		}
		if _, ok := seen[gap.ID]; ok {
			return fmt.Errorf("duplicate residual gap %s", gap.ID)
		}
		seen[gap.ID] = struct{}{}
	}
	return nil
}

func verifyResidualGap(gap residualGap, idx indexes) error {
	fields := []string{gap.ID, gap.Observation, gap.Risk, gap.NextLoop, gap.NextArtifact, gap.NextCommand}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("residual gap must bind id, observation, risk, next loop, next artifact, and next command")
		}
	}
	if _, ok := idx.loops[gap.NextLoop]; !ok {
		return fmt.Errorf("residual gap %s references unknown loop %s", gap.ID, gap.NextLoop)
	}
	return nil
}
