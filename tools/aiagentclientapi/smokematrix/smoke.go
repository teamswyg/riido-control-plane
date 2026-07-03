package smokematrix

import (
	"encoding/json"
	"os"
	"strings"
)

type matrix struct {
	Entries []entry `json:"entries"`
}

type entry struct {
	GeneratedPath string `json:"generated_path"`
}

func LoadGeneratedPaths(path string) (map[string]struct{}, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	var got matrix
	if err := json.Unmarshal(data, &got); err != nil {
		return nil, 0, err
	}
	paths := map[string]struct{}{}
	for _, entry := range got.Entries {
		paths[strings.TrimSpace(entry.GeneratedPath)] = struct{}{}
	}
	return paths, len(got.Entries), nil
}
