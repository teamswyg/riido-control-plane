package main

import "fmt"

func verifyProjection(p projectionManifest, s sourceManifest) error {
	if p.SchemaVersion != projectionSchema || p.ID != projectionID {
		return fmt.Errorf("unexpected projection identity")
	}
	if p.RiidoTask != requiredTask || p.EvidenceTool != evidenceTool {
		return fmt.Errorf("unexpected projection task or evidence tool")
	}
	if !completeLoop(p.Loop) {
		return fmt.Errorf("projection loop must be complete")
	}
	if p.Source.SchemaVersion != sourceSchema || p.Source.ID != sourceID {
		return fmt.Errorf("source pointer drifted")
	}
	if s.SchemaVersion != p.Source.SchemaVersion || s.ID != p.Source.ID {
		return fmt.Errorf("mirrored source identity drifted")
	}
	return verifyCounts(p, s)
}

func verifyCounts(p projectionManifest, s sourceManifest) error {
	if len(p.Entries) != 16 || len(p.LegacyAbsorptions) != 7 {
		return fmt.Errorf("projection coverage count drifted")
	}
	if len(p.PlanningAbsorptions) != 1 || len(p.ToolLimitations) != 3 {
		return fmt.Errorf("projection absorption or limitation count drifted")
	}
	if len(s.ExpectedPages) != 3 || len(s.NonUITopLevelNodes) != 12 {
		return fmt.Errorf("source page or non-UI coverage count drifted")
	}
	if len(s.APIGeneratedAnnotationInventory) != 20 {
		return fmt.Errorf("source API Generated inventory count drifted")
	}
	if s.AnnotationPolicy.LiveInspection.TotalRiidoAnnotations != 90 ||
		s.AnnotationPolicy.LiveInspection.TotalAPIGeneratedAnnotations != 90 {
		return fmt.Errorf("source API Generated annotation totals drifted")
	}
	return nil
}
