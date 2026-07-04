package main

type loopClaimSet struct {
	Count    int
	Observe  map[string]bool
	Verify   map[string]bool
	Fail     map[string]bool
	Evidence map[string]bool
}

func loopClaimCoverage(claims []claimBinding) map[string]loopClaimSet {
	out := map[string]loopClaimSet{}
	for _, claim := range claims {
		set := out[claim.Loop]
		set.Count++
		set.Observe = addCoverage(set.Observe, claim.CoversObserves)
		set.Verify = addCoverage(set.Verify, claim.CoversVerifies)
		set.Fail = addCoverage(set.Fail, claim.CoversFails)
		set.Evidence = addCoverage(set.Evidence, claim.CoversEvidence)
		out[claim.Loop] = set
	}
	return out
}

func addCoverage(dst map[string]bool, values []string) map[string]bool {
	if dst == nil {
		dst = map[string]bool{}
	}
	for _, value := range values {
		dst[value] = true
	}
	return dst
}

func (s loopClaimSet) CoversAll(values []string, axis string) bool {
	covered := s.coverage(axis)
	for _, value := range values {
		if !covered[value] {
			return false
		}
	}
	return len(values) > 0
}

func (s loopClaimSet) CoversEvidence(values []evidenceSource) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if !s.Evidence[value.Path] {
			return false
		}
	}
	return true
}

func (s loopClaimSet) coverage(axis string) map[string]bool {
	switch axis {
	case "observes":
		return s.Observe
	case "verifies":
		return s.Verify
	case "fails":
		return s.Fail
	default:
		return nil
	}
}
