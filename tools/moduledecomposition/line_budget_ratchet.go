package main

import "fmt"

func verifyLineBudgetRatchet(result lineBudgetResult) error {
	if result.MaxFilesOverTarget > 0 && result.OverTarget > result.MaxFilesOverTarget {
		return fmt.Errorf("line budget files over target increased: got %d, max %d",
			result.OverTarget, result.MaxFilesOverTarget)
	}
	if result.MaxFileLinesLimit > 0 && result.MaxLines > result.MaxFileLinesLimit {
		return fmt.Errorf("line budget max file lines increased: got %d, max %d",
			result.MaxLines, result.MaxFileLinesLimit)
	}
	if err := verifyLineBudgetHotspotRatchets(result.HotspotRatchets); err != nil {
		return err
	}
	return nil
}

func verifyLineBudgetHotspotRatchets(ratchets []lineBudgetHotspotRatchet) error {
	for _, ratchet := range ratchets {
		if ratchet.FilesSlack < 0 || ratchet.MaxLinesSlack < 0 || ratchet.TotalOverSlack < 0 {
			return fmt.Errorf("line budget hotspot %q increased", ratchet.Path)
		}
	}
	return nil
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
