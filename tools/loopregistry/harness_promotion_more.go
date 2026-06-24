package main

import (
	"fmt"
	"strings"
)

func harnessCandidateArtifact(loop loopRecord) (string, error) {
	for _, source := range loop.Evidence {
		if source.Kind != "workflow_artifact" || !source.Redacted {
			continue
		}
		if strings.Contains(source.Path, "closed-loop-candidate") {
			return source.Path, nil
		}
	}
	return "", fmt.Errorf("harness loop %s must declare redacted closed-loop candidate artifact evidence", loop.ID)
}

func verifyHarnessTargets(m manifest, loop loopRecord) error {
	kinds := map[string]string{}
	for _, candidate := range m.Loops {
		kinds[candidate.ID] = candidate.Kind
	}
	for _, target := range loop.PromotesTo {
		if kinds[target] != kindClosedLoop {
			return fmt.Errorf("harness loop %s promotes to non-closed-loop %s", loop.ID, target)
		}
	}
	return nil
}
