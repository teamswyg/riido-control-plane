package main

func verifyBuildModuleDownload(contract imageContract, buildStage *stage, require requireFunc) error {
	moduleDownload := contract.Build.ModuleDownload
	if err := require(hasModuleDownloadRun(buildStage.Runs, moduleDownload),
		"module download command %q not found", moduleDownload.Command); err != nil {
		return err
	}
	return requireCacheMounts(buildStage, "module download", moduleDownload.Command, moduleDownload.CacheMounts, require)
}

func verifyBuildCommand(contract imageContract, buildStage *stage, require requireFunc) error {
	if err := require(hasGoBuildRun(buildStage.Runs, contract.Build.GoBuild),
		"go build command does not satisfy container contract"); err != nil {
		return err
	}
	return requireCacheMounts(buildStage, "go build", "go build", contract.Build.GoBuild.CacheMounts, require)
}

func requireCacheMounts(buildStage *stage, label, command string, targets []string, require requireFunc) error {
	for _, target := range targets {
		if err := require(runHasCacheMount(buildStage.Runs, command, target),
			"%s cache mount target %q not found", label, target); err != nil {
			return err
		}
	}
	return nil
}
