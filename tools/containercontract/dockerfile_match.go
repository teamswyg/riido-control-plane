package main

import "strings"

func hasModuleDownloadRun(runs []string, want moduleDownloadContract) bool {
	for _, run := range runs {
		if strings.Contains(run, want.Command) {
			return true
		}
	}
	return false
}

func hasGoBuildRun(runs []string, want goBuildContract) bool {
	for _, run := range runs {
		if !strings.Contains(run, "go build") {
			continue
		}
		if !strings.Contains(run, "-o "+want.Output) || !strings.Contains(run, want.Package) {
			continue
		}
		if want.Trimpath && !strings.Contains(run, "-trimpath") {
			continue
		}
		if goBuildRunMatchesFlags(run, want.LDFlags) {
			return true
		}
	}
	return false
}

func goBuildRunMatchesFlags(run string, flags []string) bool {
	for _, flag := range flags {
		if !strings.Contains(run, flag) {
			return false
		}
	}
	return true
}

func runHasCacheMount(runs []string, command, target string) bool {
	for _, run := range runs {
		if strings.Contains(run, command) && strings.Contains(run, "type=cache,target="+target) {
			return true
		}
	}
	return false
}

func hasCopy(copies []copyInstruction, from, src, dst string) bool {
	for _, copyInstruction := range copies {
		if copyInstruction.From == from && copyInstruction.Src == src && copyInstruction.Dst == dst {
			return true
		}
	}
	return false
}
