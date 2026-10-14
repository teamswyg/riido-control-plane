package main

import "fmt"

func mergeRefreshCommandSources(sources []refreshCommandEvidence) (refreshCommandEvidence, error) {
	if len(sources) == 0 {
		return refreshCommandEvidence{}, fmt.Errorf("-commands-in is required")
	}
	commands := []selectedRefreshCommand{}
	generatedAt := ""
	expiresAt := ""
	staleSources := []staleRefreshSource{}
	freshCount := 0
	now := evidenceNow()
	for _, source := range sources {
		if err := verifySourceFresh(source, now); err != nil {
			staleSources = append(staleSources, staleRefreshSource{
				SourcePath:  source.SourcePath,
				GeneratedAt: source.GeneratedAt,
				ExpiresAt:   source.ExpiresAt,
				Reason:      err.Error(),
			})
			continue
		}
		if err := verifySourceCommands(source); err != nil {
			return refreshCommandEvidence{}, err
		}
		freshCount++
		generatedAt = latestEvidenceTime(generatedAt, source.GeneratedAt)
		expiresAt = earliestEvidenceTime(expiresAt, source.ExpiresAt)
		commands = append(commands, source.Commands...)
	}
	if freshCount == 0 {
		return refreshCommandEvidence{
			SchemaVersion: refreshCommandsSchema,
			Status:        "source_stale",
			StaleSources:  staleSources,
		}, nil
	}
	return refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        mergedRefreshStatus(commands),
		GeneratedAt:   generatedAt,
		ExpiresAt:     expiresAt,
		CommandCount:  len(commands),
		Commands:      commands,
		StaleSources:  staleSources,
	}, nil
}

func mergedRefreshStatus(commands []selectedRefreshCommand) string {
	if len(commands) == 0 {
		return "fresh"
	}
	return "refresh_required"
}
