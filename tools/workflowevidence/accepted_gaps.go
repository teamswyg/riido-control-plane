package main

func acceptedByPath(gaps []acceptedGap) map[string]acceptedGap {
	out := make(map[string]acceptedGap, len(gaps))
	for _, gap := range gaps {
		out[slashPath(gap.Path)] = gap
	}
	return out
}
