package main

import "fmt"

func writeRefreshCommandEvidence(opt options) error {
	if opt.RefreshCommandsOut == "" {
		return fmt.Errorf("-refresh-commands-out is required with -refresh-plan-in")
	}
	source, err := loadLoopRegistryEvidence(opt.RefreshPlanIn)
	if err != nil {
		return err
	}
	selected, err := selectExpiredRefreshCommands(source)
	if err != nil {
		return err
	}
	return writeJSON(opt.RefreshCommandsOut, selected)
}
