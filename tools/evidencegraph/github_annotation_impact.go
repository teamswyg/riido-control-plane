package main

import "fmt"

func impactAnnotationMessage(impact *impactEvidence) string {
	return fmt.Sprintf(
		"%d changed files, %d added chains, %d changed chains, %d removed chains",
		impact.ChangedFileCount,
		impact.AddedChainCount,
		impact.ChangedChainCount,
		impact.RemovedChainCount,
	)
}

func impactAnnotationChains(impact *impactEvidence) []impactChain {
	out := []impactChain{}
	out = append(out, impact.AddedChains...)
	out = append(out, impact.ChangedChains...)
	out = append(out, impact.RemovedChains...)
	return out
}
