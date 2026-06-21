package main

const manifestLoopSampleLimit = 3

func scanManifestLoops(root string, m manifest) manifestLoopReport {
	report := manifestLoopReport{}
	missingByGroup := map[string]int{}
	missingSamples := map[string][]string{}
	sources := manifestLoopSources(m)
	for _, path := range manifestInventory(root) {
		group := manifestGroup(path)
		switch manifestLoopStatus(root, path, sources) {
		case "direct":
			report.Complete++
			report.Direct++
			continue
		case "delegated":
			report.Complete++
			report.Delegated++
			continue
		}
		report.Missing++
		missingByGroup[group]++
		if len(missingSamples[group]) < manifestLoopSampleLimit {
			missingSamples[group] = append(missingSamples[group], path)
		}
	}
	report.MissingGroups = manifestLoopGroups(missingByGroup)
	report.MissingSamples = orderedManifestSamples(report.MissingGroups, missingSamples)
	return report
}
