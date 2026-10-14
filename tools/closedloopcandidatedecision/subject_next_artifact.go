package main

import "encoding/json"

func subjectNextArtifact(item closedLoopCandidate) (string, error) {
	if len(item.Subject) == 0 {
		return "", nil
	}
	var subject struct {
		NextArtifact string `json:"next_artifact"`
	}
	if err := json.Unmarshal(item.Subject, &subject); err != nil {
		return "", err
	}
	return subject.NextArtifact, nil
}
