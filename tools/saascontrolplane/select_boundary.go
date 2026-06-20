package main

import "fmt"

func selectBoundary(m manifest, id string) (*boundary, error) {
	if id == "" {
		return nil, nil
	}
	for i := range m.Boundaries {
		if m.Boundaries[i].ID == id {
			return &m.Boundaries[i], nil
		}
	}
	return nil, fmt.Errorf("unknown boundary %q", id)
}
