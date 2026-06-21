package main

func manifestInventorySamples(root string, limit int) []manifestSample {
	samples := map[string][]string{}
	for _, path := range manifestInventory(root) {
		group := manifestGroup(path)
		if len(samples[group]) < limit {
			samples[group] = append(samples[group], path)
		}
	}
	return orderedManifestSamples(manifestInventoryByGroup(root), samples)
}

func orderedManifestSamples(groups []manifestDir, samples map[string][]string) []manifestSample {
	ordered := make([]manifestSample, 0, len(groups))
	for _, group := range groups {
		ordered = append(ordered, manifestSample{Group: group.Group, Paths: samples[group.Group]})
	}
	return ordered
}
