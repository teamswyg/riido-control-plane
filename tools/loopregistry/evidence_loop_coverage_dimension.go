package main

func loopCoverageDimensionSurfaces() []loopCoverageDimensionSurface {
	out := make([]loopCoverageDimensionSurface, 0, len(loopCoverageDimensions))
	for _, dim := range loopCoverageDimensions {
		out = append(out, loopCoverageDimensionSurface{
			ID:              dim.id,
			LoopField:       dim.loopField,
			ClaimField:      dim.claimField,
			LoopTokenLabel:  dim.loopTokenLabel,
			ClaimTokenLabel: dim.claimTokenLabel,
		})
	}
	return out
}
