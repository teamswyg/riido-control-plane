package main

import (
	"fmt"
	"strings"
)

var forbiddenCandidateFragments = []string{
	"http://",
	"https://",
	"arn:aws:",
	"AKIA",
	"ASIA",
	"eyJ",
	"staging.ai-api.riido.io",
	"development.ai-api.riido.io",
	"prod.ai-api.riido.io",
}

func verifyNoRawLeak(data []byte) error {
	text := string(data)
	for _, fragment := range forbiddenCandidateFragments {
		if strings.Contains(text, fragment) {
			return fmt.Errorf("candidate artifact contains forbidden raw fragment %q", fragment)
		}
	}
	return nil
}
