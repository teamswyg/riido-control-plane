package main

import (
	"encoding/json"
	"os"
)

func loadRefreshCommandEvidence(path string) (refreshCommandEvidence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return refreshCommandEvidence{}, err
	}
	var got refreshCommandEvidence
	if err := json.Unmarshal(data, &got); err != nil {
		return refreshCommandEvidence{}, err
	}
	return got, nil
}
