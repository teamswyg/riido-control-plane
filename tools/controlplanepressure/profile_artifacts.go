package main

import (
	"os"
	"path/filepath"
)

func profileArtifacts(dir string) ([]profileArtifact, error) {
	specs := []struct {
		kind string
		file string
	}{
		{"cpu", "cpu.pprof"},
		{"heap", "heap.pprof"},
		{"goroutine", "goroutine.txt"},
	}
	out := make([]profileArtifact, 0, len(specs))
	for _, spec := range specs {
		path := filepath.Join(dir, spec.file)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		out = append(out, profileArtifact{Kind: spec.kind, Path: path, Bytes: info.Size()})
	}
	return out, nil
}
