package main

import "fmt"

func testWorkflows() []string {
	out := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		out = append(out, fmt.Sprintf(".github/workflows/w%02d.yml", i))
	}
	return out
}

func testBoundaries(workflows []string) []boundary {
	out := make([]boundary, 0, 12)
	for i := 0; i < 12; i++ {
		id := fmt.Sprintf("b%02d", i)
		out = append(out, boundary{
			ID:               id,
			Summary:          "Boundary " + id,
			Workflow:         workflows[i%len(workflows)],
			EvidenceArtifact: "artifact-" + id,
			SourceChecks: []sourceCheck{{
				Name:     "anchor",
				File:     "anchors/" + id + ".txt",
				Contains: []string{"anchor-" + id},
			}},
		})
	}
	return out
}

func testArtifactText(m manifest) string {
	out := ""
	for _, item := range m.Boundaries {
		out += item.EvidenceArtifact + "\nevidence-out\n"
	}
	return out
}
