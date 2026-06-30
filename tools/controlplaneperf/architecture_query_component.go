package main

import "sort"

func architectureQueryComponents(
	ids []string,
	components []architectureComponent,
) []architectureQueryComponent {
	byID := map[string]string{}
	for _, component := range components {
		byID[component.ID] = component.Role
	}
	out := make([]architectureQueryComponent, 0, len(ids))
	for _, id := range ids {
		out = append(out, architectureQueryComponent{ID: id, Role: byID[id]})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}
