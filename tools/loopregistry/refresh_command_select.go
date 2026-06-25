package main

import (
	"fmt"
	"time"
)

func selectExpiredRefreshCommands(source evidence) (refreshCommandEvidence, error) {
	now := evidenceNow()
	out := refreshCommandEvidence{
		SchemaVersion:     refreshCommandsSchema,
		Status:            "fresh",
		GeneratedAt:       formatEvidenceTime(now),
		SourceGeneratedAt: source.GeneratedAt,
		SourceExpiresAt:   source.ExpiresAt,
	}
	for _, plan := range source.RefreshPlans {
		expired, err := refreshPlanExpired(plan, source.ExpiresAt, now)
		if err != nil {
			return refreshCommandEvidence{}, err
		}
		if !expired {
			continue
		}
		out.SelectedLoops = append(out.SelectedLoops, selectedLoop(plan))
		for _, command := range plan.NextCommands {
			out.Commands = append(out.Commands, selectedCommand(plan.LoopID, command))
		}
	}
	out.SelectedLoopCount = len(out.SelectedLoops)
	out.CommandCount = len(out.Commands)
	if out.CommandCount > 0 {
		out.Status = "refresh_required"
	}
	return out, nil
}

func refreshPlanExpired(plan refreshPlan, fallback string, now time.Time) (bool, error) {
	value := plan.EvidenceExpiresAt
	if value == "" {
		value = fallback
	}
	if value == "" {
		return false, fmt.Errorf("refresh plan %s missing evidence expiry", plan.LoopID)
	}
	expiresAt, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false, fmt.Errorf("refresh plan %s invalid evidence expiry: %w", plan.LoopID, err)
	}
	return !now.Before(expiresAt), nil
}
