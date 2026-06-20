package main

import (
	"os"
	"strings"
)

func parseDockerfile(path string) (dockerfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return dockerfile{}, err
	}
	defer file.Close()
	var parsed dockerfile
	parsed.Args = map[string]string{}
	var current *stage
	for _, logical := range logicalDockerfileLines(file) {
		instruction, rest, ok := splitInstruction(logical)
		if !ok {
			continue
		}
		current = applyDockerfileInstruction(&parsed, current, instruction, rest)
	}
	return parsed, nil
}

func applyDockerfileInstruction(parsed *dockerfile, current *stage, instruction, rest string) *stage {
	switch instruction {
	case "ARG":
		name, value := splitKeyValue(rest)
		parsed.Args[name] = value
	case "FROM":
		base, alias := parseFrom(rest)
		parsed.Stages = append(parsed.Stages, stage{Base: base, Alias: alias, Env: map[string]string{}})
		current = &parsed.Stages[len(parsed.Stages)-1]
	case "WORKDIR":
		if current != nil {
			current.Workdir = strings.TrimSpace(rest)
		}
	case "ENV":
		applyEnv(current, rest)
	case "RUN":
		applyRun(current, rest)
	case "COPY":
		applyCopy(current, rest)
	case "EXPOSE":
		applyExpose(current, rest)
	case "USER":
		applyUser(current, rest)
	case "ENTRYPOINT":
		applyEntrypoint(current, rest)
	}
	return current
}
