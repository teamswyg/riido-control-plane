package main

import (
	"fmt"
	"os"
	"time"
)

func readinessNow() (time.Time, error) {
	value := os.Getenv(readinessNowEnv)
	if value == "" {
		return time.Now().UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s is invalid: %w", readinessNowEnv, err)
	}
	return parsed.UTC(), nil
}
