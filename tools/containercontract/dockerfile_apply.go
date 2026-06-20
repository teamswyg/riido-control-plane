package main

import "strings"

func applyEnv(current *stage, rest string) {
	if current == nil {
		return
	}
	name, value := splitKeyValue(rest)
	current.Env[name] = value
}

func applyRun(current *stage, rest string) {
	if current != nil {
		current.Runs = append(current.Runs, strings.TrimSpace(rest))
	}
}

func applyCopy(current *stage, rest string) {
	if current != nil {
		current.Copies = append(current.Copies, parseCopy(rest))
	}
}

func applyExpose(current *stage, rest string) {
	if current != nil {
		current.Exposes = append(current.Exposes, parseExposedPorts(rest)...)
	}
}

func applyUser(current *stage, rest string) {
	if current != nil {
		current.User = strings.TrimSpace(rest)
	}
}

func applyEntrypoint(current *stage, rest string) {
	if current != nil {
		current.Entrypoint = parseExecJSON(rest)
	}
}
