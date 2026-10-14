package main

import (
	"fmt"
	"strings"
)

func verifySourceCommands(source refreshCommandEvidence) error {
	status := strings.TrimSpace(source.Status)
	commandCount := len(source.Commands)
	if source.CommandCount != commandCount {
		return fmt.Errorf("refresh command evidence command_count=%d does not match commands=%d",
			source.CommandCount, commandCount)
	}
	switch status {
	case "fresh":
		if commandCount != 0 {
			return fmt.Errorf("fresh refresh command evidence must not carry commands")
		}
	case "refresh_required":
		if commandCount == 0 {
			return fmt.Errorf("refresh_required evidence must carry at least one command")
		}
	default:
		return fmt.Errorf("unsupported refresh command evidence status %q", status)
	}
	return nil
}
