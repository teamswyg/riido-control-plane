package main

import "sort"

func harnessLoopIDs(loops []registryLoop) []string {
	ids := make([]string, 0, len(loops))
	for _, loop := range loops {
		ids = append(ids, loop.ID)
	}
	sort.Strings(ids)
	return ids
}

func allHarnessCandidateArtifacts(loops []registryLoop) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, loop := range loops {
		for _, artifact := range harnessCandidateArtifacts(loop) {
			if seen[artifact] {
				continue
			}
			seen[artifact] = true
			out = append(out, artifact)
		}
	}
	sort.Strings(out)
	return out
}
