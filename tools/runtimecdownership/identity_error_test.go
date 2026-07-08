package main

import "testing"

func TestVerifyIdentityRejectsManifestDrift(t *testing.T) {
	t.Parallel()
	cases := map[string]func(manifest) manifest{
		"identity": func(m manifest) manifest {
			m.ID = "other"
			return m
		},
		"target": func(m manifest) manifest {
			m.Runtime = "other"
			return m
		},
		"reader": func(m manifest) manifest {
			m.GeneratedDoc = "other.md"
			return m
		},
		"lineage": func(m manifest) manifest {
			m.Hardening = nil
			return m
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := verifyIdentity(mutate(testManifest(t))); err == nil {
				t.Fatalf("verifyIdentity accepted %s drift", name)
			}
		})
	}
}

func TestVerifyDocAndLoopRejectDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := verifyDoc(root, "expected"); err == nil {
		t.Fatal("verifyDoc accepted missing generated doc")
	}
	if _, err := verifyLoop(evidenceLoop{Observation: "o"}); err == nil {
		t.Fatal("verifyLoop accepted incomplete loop")
	}
}
