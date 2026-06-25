package main

import (
	"os"
	"strings"
)

var signalPatterns = []string{
	"sync.Mutex", "sync.RWMutex", "Lock()", "RLock()",
	"make(chan", "chan ", "go func", "select {",
	"time.NewTimer", "time.NewTicker", "map[", "append(",
	"GetItem", "PutItem", "Query", "TransactWriteItems",
}

func scanSurfaces(root string, surfaces []surface) ([]surfaceEvidence, error) {
	rows := make([]surfaceEvidence, 0, len(surfaces))
	for _, surface := range surfaces {
		files, err := scanFiles(root, surface.Files)
		if err != nil {
			return nil, err
		}
		rows = append(rows, surfaceEvidence{
			ID: surface.ID, Category: surface.Category, Risk: surface.Risk,
			Files: files, Candidate: surface.Candidate,
		})
	}
	return rows, nil
}

func scanFiles(root string, paths []string) ([]fileEvidence, error) {
	rows := make([]fileEvidence, 0, len(paths))
	for _, path := range paths {
		text, err := readText(repoPath(root, path))
		if err != nil {
			return nil, err
		}
		rows = append(rows, fileEvidence{Path: path, Signals: countSignals(text)})
	}
	return rows, nil
}

func countSignals(text string) map[string]int {
	counts := map[string]int{}
	for _, pattern := range signalPatterns {
		if count := strings.Count(text, pattern); count > 0 {
			counts[pattern] = count
		}
	}
	return counts
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
