package main

import "fmt"

func verifyLineBudgetRatchet(result lineBudgetResult) error {
	if result.MaxFileLinesLimit > 0 && result.MaxLines > result.MaxFileLinesLimit {
		return fmt.Errorf("line budget max file lines increased: got %d, max %d",
			result.MaxLines, result.MaxFileLinesLimit)
	}
	if err := verifyLineBudgetHotspotRatchets(result.HotspotRatchets); err != nil {
		return err
	}
	if err := verifyLineBudgetHotspotCoverage(result.UntrackedHotspots); err != nil {
		return err
	}
	return nil
}

func verifyLineBudgetHotspotRatchets(ratchets []lineBudgetHotspotRatchet) error {
	for _, ratchet := range ratchets {
		if ratchet.MaxLinesSlack < 0 || ratchet.TotalOverSlack < 0 {
			return fmt.Errorf("line budget hotspot %q increased", ratchet.Path)
		}
	}
	return nil
}

func verifyLineBudgetHotspotCoverage(untracked []lineBudgetHotspot) error {
	if len(untracked) == 0 {
		return nil
	}
	return fmt.Errorf("line budget untracked hotspot %q", untracked[0].Path)
}

func lineBudgetFilesSlack(result lineBudgetResult) int {
	if result.MaxFilesOverTarget <= 0 {
		return 0
	}
	return result.MaxFilesOverTarget - result.OverTarget
}

func lineBudgetMaxLinesSlack(result lineBudgetResult) int {
	if result.MaxFileLinesLimit <= 0 {
		return 0
	}
	return result.MaxFileLinesLimit - result.MaxLines
}
