package main

import "strings"

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}

func completeLoop(loop evidenceLoop) bool {
	for _, value := range []string{
		loop.Observation, loop.Hypothesis, loop.Execute,
		loop.Evaluate, loop.Retrospective,
	} {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}
