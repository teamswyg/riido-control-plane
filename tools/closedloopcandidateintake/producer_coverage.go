package main

import "fmt"

func verifyProducerCoverage(root string, m manifest) error {
	for manifestPath, sources := range sourcesByProducerManifest(m.Sources) {
		var producer producerManifest
		if err := readJSON(repoPath(root, manifestPath), &producer); err != nil {
			return err
		}
		if err := verifyProducerSourcesCovered(manifestPath, producer.Sources, sources); err != nil {
			return err
		}
	}
	return nil
}

func sourcesByProducerManifest(sources []intakeSource) map[string][]intakeSource {
	byManifest := map[string][]intakeSource{}
	for _, source := range sources {
		byManifest[source.ProducerManifest] = append(byManifest[source.ProducerManifest], source)
	}
	return byManifest
}

func verifyProducerSourcesCovered(path string, producers []producerSource, sources []intakeSource) error {
	byID := intakeSourcesByID(sources)
	for _, producer := range producers {
		if producer.PromotionTarget != "closed_loop_candidate" {
			continue
		}
		source, ok := byID[producer.ID]
		if !ok {
			return fmt.Errorf("producer source %s from %s has no intake source", producer.ID, path)
		}
		if err := verifyProducerSource(source, producer); err != nil {
			return err
		}
	}
	return nil
}

func intakeSourcesByID(sources []intakeSource) map[string]intakeSource {
	byID := map[string]intakeSource{}
	for _, source := range sources {
		byID[source.ID] = source
	}
	return byID
}
