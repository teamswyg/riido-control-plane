package main

func loopSurfaces(loops []loopRecord) []loopSurface {
	out := make([]loopSurface, 0, len(loops))
	for _, loop := range loops {
		out = append(out, loopSurface{
			ID:                loop.ID,
			Kind:              loop.Kind,
			Observes:          append([]string(nil), loop.Observes...),
			Verifies:          append([]string(nil), loop.Verifies...),
			Evidence:          append([]evidenceSource(nil), loop.Evidence...),
			RefreshWorkflow:   loop.RefreshWorkflow,
			ExpiresAfterHours: loop.ExpiresAfterHours,
			FailsWhen:         append([]string(nil), loop.FailsWhen...),
			PromotesTo:        append([]string(nil), loop.PromotesTo...),
			Providers:         append([]string(nil), loop.Providers...),
		})
	}
	return out
}
