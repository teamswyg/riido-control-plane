package main

import "strings"

func requirementProofs(checks []check, idxOpt ...indexes) []proof {
	out := make([]proof, 0, len(checks))
	for _, c := range checks {
		row := proof{
			Kind:   c.Kind,
			Key:    proofKey(c),
			Status: "verified",
		}
		if len(idxOpt) > 0 {
			row.Surface = proofSurfaceFor(c, idxOpt[0])
		}
		out = append(out, row)
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
