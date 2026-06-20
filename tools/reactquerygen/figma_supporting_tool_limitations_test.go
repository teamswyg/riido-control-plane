package main

import "testing"

type figmaSupportingLimitationContext struct {
	sourceByID     map[string]figmaSourceSupportingToolLimitation
	projectionByID map[string]figmaProjectionSupportingToolLimitation
	stabilizedBy   []string
	docText        string
}

func verifyMirroredFigmaSupportingToolLimitations(t *testing.T, projections []figmaProjectionSupportingToolLimitation, sourceLimitations []figmaSourceSupportingToolLimitation, stabilizedBy []string, docText string) {
	t.Helper()
	if len(projections) == 0 {
		t.Fatalf("mirrored_supporting_tool_limitations must record consumed source tooling limits")
	}
	ctx := figmaSupportingLimitationContext{
		sourceByID:     figmaSourceLimitationsByID(sourceLimitations),
		projectionByID: figmaProjectionLimitationsByID(projections),
		stabilizedBy:   stabilizedBy,
		docText:        docText,
	}
	verifyMetadataPageListLimitation(t, ctx)
	verifyHeadlessFileKeyLimitation(t, ctx)
	verifyOnboardingTimeoutLimitation(t, ctx)
}

func figmaSourceLimitationsByID(limitations []figmaSourceSupportingToolLimitation) map[string]figmaSourceSupportingToolLimitation {
	out := map[string]figmaSourceSupportingToolLimitation{}
	for _, limitation := range limitations {
		out[limitation.ID] = limitation
	}
	return out
}

func figmaProjectionLimitationsByID(limitations []figmaProjectionSupportingToolLimitation) map[string]figmaProjectionSupportingToolLimitation {
	out := map[string]figmaProjectionSupportingToolLimitation{}
	for _, limitation := range limitations {
		out[limitation.SourceID] = limitation
	}
	return out
}

func requireFigmaSupportingLimitation(t *testing.T, ctx figmaSupportingLimitationContext, id string) (figmaProjectionSupportingToolLimitation, figmaSourceSupportingToolLimitation) {
	t.Helper()
	projection, ok := ctx.projectionByID[id]
	if !ok {
		t.Fatalf("mirrored_supporting_tool_limitations must include %s", id)
	}
	source, ok := ctx.sourceByID[id]
	if !ok {
		t.Fatalf("projection supporting limit %q is missing from mirrored contracts coverage", id)
	}
	return projection, source
}
