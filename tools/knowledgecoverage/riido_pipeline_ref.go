package main

import "fmt"

type riidoPipelineRef struct {
	Path          string `json:"path"`
	OwnerManifest string `json:"owner_manifest"`
	OwnerKey      string `json:"owner_key"`
}

func validateRiidoPipelines(root string, m manifest) []string {
	var problems []string
	for _, pipeline := range m.RiidoPipelines {
		problems = append(problems, validateRiidoPipeline(root, m, pipeline)...)
	}
	return problems
}

func validateRiidoPipeline(root string, m manifest, ref riidoPipelineRef) []string {
	if ref.Path == "" || ref.OwnerManifest == "" || ref.OwnerKey == "" {
		return []string{"riido-ci pipeline path, owner_manifest, and owner_key are required"}
	}
	if !ownerManifestDeclaresPath(root, ref.OwnerManifest, ref.OwnerKey, ref.Path) {
		return []string{fmt.Sprintf("%s owner manifest %q key %q must declare pipeline path",
			ref.Path, ref.OwnerManifest, ref.OwnerKey)}
	}
	pipeline, ok := readRiidoPipeline(root, ref.Path)
	if !ok {
		return []string{fmt.Sprintf("%s must be an active attested private riido-ci pipeline", ref.Path)}
	}
	for _, source := range m.SourceManifests {
		if source.Path == ref.OwnerManifest && source.Workflow == ref.Path &&
			source.EvidenceArtifact == pipeline.Evidence.Artifact &&
			len(validateSourceManifest(root, source)) == 0 {
			return nil
		}
	}
	return []string{fmt.Sprintf("%s must be owned by a strict source-manifest evidence route", ref.Path)}
}
