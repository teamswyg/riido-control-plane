package main

import "fmt"

func verifyGraph(m manifest, graph syntaxGraph) error {
	if graph.Repository.CoverageBasisPoints < m.Constraints.MinRepositoryCoverageBasisPoint {
		return fmt.Errorf("repository coverage %d bp below %d bp",
			graph.Repository.CoverageBasisPoints, m.Constraints.MinRepositoryCoverageBasisPoint)
	}
	for i, target := range graph.Targets {
		if m.Targets[i].GoldenCommand == "" {
			return fmt.Errorf("target %s missing golden command", target.ID)
		}
		if target.Coverage < m.Constraints.MinCoveragePercent {
			return fmt.Errorf("target %s coverage %d below %d",
				target.ID, target.Coverage, m.Constraints.MinCoveragePercent)
		}
		if target.PackageHash == "" || target.SemanticHash == "" {
			return fmt.Errorf("target %s missing syntax or semantic hash", target.ID)
		}
		if target.GoFiles != target.TrackedFiles {
			return fmt.Errorf("target %s has untracked go files", target.ID)
		}
		if len(target.Relocations) != len(target.FileHashes) {
			return fmt.Errorf("target %s missing relocation mapping", target.ID)
		}
	}
	return nil
}
