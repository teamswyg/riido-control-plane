package main

const evidenceGraphManifest = "docs/30-architecture/evidence-graph.riido.json"

func changedValues(values []string, changed map[string]bool) []string {
	out := []string{}
	for _, value := range values {
		if changed[value] {
			out = append(out, value)
		}
	}
	return out
}

func prefixedValues(prefix string, values []string) []string {
	out := []string{}
	for _, value := range sortedCopy(values) {
		out = append(out, prefix+":"+value)
	}
	return out
}
