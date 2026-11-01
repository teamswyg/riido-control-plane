package riidoaiserver

import "strings"

type cliAgentReadTarget struct {
	workspace, task, agent string
	page                   int
}

func cliAgentReadPath(path string) (cliAgentReadTarget, bool) {
	parts := strings.Split(path, "/")
	if len(parts) != 8 && len(parts) != 9 {
		return cliAgentReadTarget{}, false
	}
	if parts[0] != "" || parts[1] != "v1" || parts[2] != "cli" || parts[3] != "workspaces" || parts[5] != "tasks" || parts[7] != "agents" || !cliAgentReadID(parts[4]) || !cliAgentReadID(parts[6]) {
		return cliAgentReadTarget{}, false
	}
	target := cliAgentReadTarget{workspace: parts[4], task: parts[6]}
	if len(parts) == 9 {
		if !cliAgentReadID(parts[8]) {
			return cliAgentReadTarget{}, false
		}
		target.agent = parts[8]
	}
	return target, true
}

func cliAgentReadID(value string) bool {
	if value == "" || len(value) > 512 {
		return false
	}
	for _, c := range value {
		if c == '-' || c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}
