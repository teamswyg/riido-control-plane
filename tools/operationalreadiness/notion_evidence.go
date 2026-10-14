package main

func newNotionEvidence(openLoop *notionOpenLoop) notionEvidence {
	if openLoop == nil {
		return notionEvidence{}
	}
	statusCounts := map[string]int{}
	codexStatuses := map[string]int{}
	cycles := make([]notionCycleEvidence, 0, len(openLoop.Cycles))
	p0Count, partialCount, coveredCount := 0, 0, 0
	for _, cycle := range openLoop.Cycles {
		statusCounts[cycle.Status]++
		codexStatuses[cycle.CodexStatus]++
		if cycle.Priority == "P0" {
			p0Count++
		}
		if cycle.Status == "partial" {
			partialCount++
		}
		if cycle.Status == "covered" {
			coveredCount++
		}
		cycles = append(cycles, notionCycleEvidence{
			ID: cycle.ID, Priority: cycle.Priority, Status: cycle.Status,
			CodexStatus: cycle.CodexStatus, Source: cycle.Source,
			BackfilledCheck:      cycle.BackfilledCheck,
			RequiredNextArtifact: cycle.RequiredNextArtifact,
		})
	}
	return notionEvidence{
		PageTitle: openLoop.PageTitle, PageURL: openLoop.PageURL,
		CapturedAt: openLoop.CapturedAt, CadenceHours: openLoop.CadenceHours,
		CycleCount: len(openLoop.Cycles), P0Count: p0Count,
		PartialCount: partialCount, CoveredCount: coveredCount,
		StatusCounts: statusCounts, CodexStatuses: codexStatuses, Cycles: cycles,
	}
}
