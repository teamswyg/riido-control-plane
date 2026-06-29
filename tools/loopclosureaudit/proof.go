package main

import "strings"

type proof struct {
	Kind   string `json:"kind"`
	Key    string `json:"key"`
	Status string `json:"status"`
}

func requirementProofs(checks []check) []proof {
	out := make([]proof, 0, len(checks))
	for _, c := range checks {
		out = append(out, proof{
			Kind:   c.Kind,
			Key:    proofKey(c),
			Status: "verified",
		})
	}
	return out
}

func proofKey(c check) string {
	switch c.Kind {
	case "loop", "claim", "graph_chain", "pre_commit_hook":
		return c.Kind + ":" + c.ID
	case "workflow":
		return "workflow:" + c.Path + ":" + strings.Join(c.Contains, ",")
	case "graph_edge":
		return "graph_edge:" + c.From + ":" + c.Relation + ":" + c.To
	default:
		return c.Kind
	}
}
