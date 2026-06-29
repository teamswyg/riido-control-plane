package main

import "testing"

func findProof(t *testing.T, e evidence, key string) proof {
	t.Helper()
	for _, req := range e.Requirements {
		for _, proof := range req.Proofs {
			if proof.Key == key {
				return proof
			}
		}
	}
	t.Fatalf("missing proof %s", key)
	return proof{}
}
