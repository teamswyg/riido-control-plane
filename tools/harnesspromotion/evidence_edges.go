package main

import "sort"

func promotionEdgesForSources(sources []promotionSource) []graphEdge {
	seen := map[graphEdge]struct{}{}
	for _, source := range sources {
		edge := promotionEdge(source)
		seen[edge] = struct{}{}
	}
	edges := make([]graphEdge, 0, len(seen))
	for edge := range seen {
		edges = append(edges, edge)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		return edges[i].Relation < edges[j].Relation
	})
	return edges
}
