package main

import (
	"fmt"
	"os"
	"strings"
)

const envTaskContextLocatorVersion = "RIIDO_AI_SERVER_TASK_CONTEXT_LOCATOR_VERSION"

func taskContextLocatorVersionFromEnv() (int, error) {
	switch strings.TrimSpace(os.Getenv(envTaskContextLocatorVersion)) {
	case "", "1":
		return 1, nil
	case "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("%s must be 1 or 2", envTaskContextLocatorVersion)
	}
}
