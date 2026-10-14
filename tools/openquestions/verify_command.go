package main

import (
	"fmt"
	"strings"
)

var allowedNextCommandPrefixes = []string{
	"gh issue create ",
	"gh workflow run ",
	"go run ",
}

func verifyNextCommand(item question) error {
	if strings.TrimSpace(item.NextCommand) == "" || item.NextCommand == "none" {
		return fmt.Errorf("open question %s requires a next command", item.ID)
	}
	for _, prefix := range allowedNextCommandPrefixes {
		if strings.HasPrefix(item.NextCommand, prefix) {
			return nil
		}
	}
	return fmt.Errorf("open question %s next command must use an allowed executable prefix", item.ID)
}
