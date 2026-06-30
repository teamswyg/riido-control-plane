package main

import "strings"

func idProofKey(c check) string {
	return c.Kind + ":" + c.ID
}

func workflowProofKey(c check) string {
	return "workflow:" + c.Path + ":" + strings.Join(c.Contains, ",")
}

func graphEdgeProofKey(c check) string {
	return "graph_edge:" + c.From + ":" + c.Relation + ":" + c.To
}
