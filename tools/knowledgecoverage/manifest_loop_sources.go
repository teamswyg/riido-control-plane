package main

func manifestLoopSources(m manifest) map[string]string {
	sources := map[string]string{}
	for _, owned := range m.OwnedManifests {
		sources[owned.Path] = owned.OwnerManifest
	}
	for _, artifact := range m.ContractArtifacts {
		sources[artifact.Path] = artifact.OwnerManifest
	}
	for _, imported := range m.ImportedManifests {
		sources[imported.Path] = imported.OwnerManifest
	}
	for _, pipeline := range m.RiidoPipelines {
		sources[pipeline.Path] = pipeline.OwnerManifest
	}
	return sources
}
