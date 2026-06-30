package main

type sourceCoverage struct {
	SidecarSourceCount              int
	LoopOwnedCandidateProducerCount int
}

func sourceCoverageSummary(root string, registry loopRegistry) (sourceCoverage, error) {
	var out sourceCoverage
	for _, loop := range registry.Loops {
		if loop.Kind != "harness" {
			continue
		}
		uses, err := loopUsesHarnessPromotion(root, loop)
		if err != nil {
			return out, err
		}
		if uses {
			out.SidecarSourceCount++
			continue
		}
		out.LoopOwnedCandidateProducerCount++
	}
	return out, nil
}
