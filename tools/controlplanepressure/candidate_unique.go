package main

func uniqueCandidates(runs []pressureRun) []candidateEntry {
	seen := map[string]bool{}
	out := []candidateEntry{}
	for _, run := range runs {
		if seen[run.Candidate.ID] {
			continue
		}
		seen[run.Candidate.ID] = true
		out = append(out, run.Candidate)
	}
	return out
}
