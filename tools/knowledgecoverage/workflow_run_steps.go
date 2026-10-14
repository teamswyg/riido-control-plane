package main

import "strings"

func workflowRunCommands(text string) []string {
	lines := strings.Split(text, "\n")
	var commands []string
	for i := 0; i < len(lines); i++ {
		value, ok := workflowRunValue(strings.TrimSpace(lines[i]))
		if !ok {
			continue
		}
		if workflowRunValueIsBlock(value) {
			command, next := workflowRunBlock(lines, i+1, leadingSpaces(lines[i]))
			commands = append(commands, command)
			i = next - 1
			continue
		}
		commands = append(commands, strings.Trim(strings.TrimSpace(value), `"'`))
	}
	return commands
}

func workflowRunValue(trimmed string) (string, bool) {
	if value, ok := strings.CutPrefix(trimmed, "run:"); ok {
		return strings.TrimSpace(value), true
	}
	if value, ok := strings.CutPrefix(trimmed, "- run:"); ok {
		return strings.TrimSpace(value), true
	}
	return "", false
}

func workflowRunValueIsBlock(value string) bool {
	return value == "|" || value == "|-" || value == ">" || value == ">-"
}

func workflowRunBlock(lines []string, start, runIndent int) (string, int) {
	var block []string
	i := start
	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			block = append(block, "")
			continue
		}
		if leadingSpaces(lines[i]) <= runIndent {
			break
		}
		block = append(block, strings.TrimSpace(lines[i]))
	}
	return strings.Join(block, "\n"), i
}

func leadingSpaces(line string) int {
	count := 0
	for _, r := range line {
		if r != ' ' {
			return count
		}
		count++
	}
	return count
}
