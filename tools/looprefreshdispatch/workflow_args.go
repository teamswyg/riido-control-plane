package main

import "fmt"

func verifiedWorkflowArgs(args []string) ([]string, []workflowInput, error) {
	verified := []string{}
	inputs := []workflowInput{}
	refSeen := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--ref":
			if refSeen || i+1 >= len(args) || args[i+1] != "main" {
				return nil, nil, fmt.Errorf("unsupported refresh workflow ref")
			}
			refSeen = true
			i++
		case "-f":
			if i+1 >= len(args) || !safeWorkflowInput(args[i+1]) {
				return nil, nil, fmt.Errorf("unsupported refresh workflow input")
			}
			verified = append(verified, "-f", args[i+1])
			inputs = append(inputs, workflowInputFromArg(args[i+1]))
			i++
		default:
			return nil, nil, fmt.Errorf("unsupported refresh workflow argument %q", args[i])
		}
	}
	return verified, inputs, nil
}

func safeWorkflowInput(value string) bool {
	if value == "" {
		return false
	}
	seenEquals := false
	for _, r := range value {
		if r == '=' {
			seenEquals = true
			continue
		}
		if isWorkflowInputRune(r) {
			continue
		}
		return false
	}
	return seenEquals && value[0] != '=' && value[len(value)-1] != '='
}

func isWorkflowInputRune(r rune) bool {
	if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
		return true
	}
	switch r {
	case '_', '-', '.', '/', ':', '@':
		return true
	default:
		return false
	}
}
