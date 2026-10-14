package dockerfile

import (
	"os"
	"strings"
)

func Parse(path string) (File, error) {
	file, err := os.Open(path)
	if err != nil {
		return File{}, err
	}
	defer file.Close()
	parsed := File{Args: map[string]string{}}
	var current *Stage
	for _, logical := range logicalLines(file) {
		instruction, rest, ok := splitInstruction(logical)
		if ok {
			current = applyInstruction(&parsed, current, instruction, rest)
		}
	}
	return parsed, nil
}

func applyInstruction(parsed *File, current *Stage, instruction, rest string) *Stage {
	switch instruction {
	case "ARG":
		name, value := splitKeyValue(rest)
		parsed.Args[name] = value
	case "FROM":
		base, alias := parseFrom(rest)
		parsed.Stages = append(parsed.Stages, Stage{Base: base, Alias: alias, Env: map[string]string{}})
		current = &parsed.Stages[len(parsed.Stages)-1]
	case "WORKDIR":
		if current != nil {
			current.Workdir = strings.TrimSpace(rest)
		}
	case "ENV":
		applyEnv(current, rest)
	case "RUN":
		if current != nil {
			current.Runs = append(current.Runs, strings.TrimSpace(rest))
		}
	case "COPY":
		if current != nil {
			current.Copies = append(current.Copies, parseCopy(rest))
		}
	case "EXPOSE":
		if current != nil {
			current.Exposes = append(current.Exposes, parseExposedPorts(rest)...)
		}
	case "USER":
		if current != nil {
			current.User = strings.TrimSpace(rest)
		}
	case "ENTRYPOINT":
		if current != nil {
			current.Entrypoint = parseExecJSON(rest)
		}
	}
	return current
}

func applyEnv(current *Stage, rest string) {
	if current != nil {
		name, value := splitKeyValue(rest)
		current.Env[name] = value
	}
}
