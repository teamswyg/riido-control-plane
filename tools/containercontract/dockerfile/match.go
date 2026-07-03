package dockerfile

import "strings"

func HasModuleDownloadRun(runs []string, command string) bool {
	for _, run := range runs {
		if strings.Contains(run, command) {
			return true
		}
	}
	return false
}

func HasGoBuildRun(runs []string, output, pkg string, trimpath bool, flags []string) bool {
	for _, run := range runs {
		if !strings.Contains(run, "go build") {
			continue
		}
		if !strings.Contains(run, "-o "+output) || !strings.Contains(run, pkg) {
			continue
		}
		if trimpath && !strings.Contains(run, "-trimpath") {
			continue
		}
		if runMatchesFlags(run, flags) {
			return true
		}
	}
	return false
}

func runMatchesFlags(run string, flags []string) bool {
	for _, flag := range flags {
		if !strings.Contains(run, flag) {
			return false
		}
	}
	return true
}

func RunHasCacheMount(runs []string, command, target string) bool {
	for _, run := range runs {
		if strings.Contains(run, command) && strings.Contains(run, "type=cache,target="+target) {
			return true
		}
	}
	return false
}

func HasCopy(copies []CopyInstruction, from, src, dst string) bool {
	for _, copyInstruction := range copies {
		if copyInstruction.From == from && copyInstruction.Src == src && copyInstruction.Dst == dst {
			return true
		}
	}
	return false
}
