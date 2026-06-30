package main

import (
	"fmt"
	"strings"
)

func verifyNotionOpenLoop(root string, m manifest) error {
	if m.NotionOpenLoop == nil {
		return fmt.Errorf("readiness manifest must bind notion open loop")
	}
	loop := *m.NotionOpenLoop
	if loop.PageID == "" || loop.PageTitle == "" || !strings.HasPrefix(loop.PageURL, "https://") {
		return fmt.Errorf("notion open loop must bind page identity")
	}
	if loop.RefreshWorkflow != m.Workflow || loop.CadenceHours != readinessEvidenceTTLHours {
		return fmt.Errorf("notion open loop must refresh with readiness workflow cadence")
	}
	if len(loop.CapturedAt) < 10 {
		return fmt.Errorf("notion open loop captured_at must include date")
	}
	if _, err := readinessDate(loop.CapturedAt[:10]); err != nil {
		return fmt.Errorf("notion open loop captured_at must start with date: %w", err)
	}
	return verifyNotionCycles(root, m, loop.Cycles)
}

func verifyNotionCycles(root string, m manifest, cycles []notionCycle) error {
	if len(cycles) == 0 {
		return fmt.Errorf("notion open loop must bind p0 cycles")
	}
	checks := readinessCheckIDs(m)
	seen := map[string]bool{}
	for _, cycle := range cycles {
		if err := verifyNotionCycle(root, checks, cycle); err != nil {
			return err
		}
		if seen[cycle.ID] {
			return fmt.Errorf("duplicate notion cycle %s", cycle.ID)
		}
		seen[cycle.ID] = true
	}
	return nil
}

func readinessCheckIDs(m manifest) map[string]bool {
	ids := map[string]bool{}
	for _, check := range m.Checks {
		ids[check.ID] = true
	}
	return ids
}
