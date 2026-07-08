package main

import "testing"

func TestVerifyInfraBoundaryRejectsDrift(t *testing.T) {
	t.Parallel()
	cases := map[string]func(manifest) manifest{
		"repo": func(m manifest) manifest {
			m.Infra.Repo = "other"
			return m
		},
		"task": func(m manifest) manifest {
			m.InfraTopology.RiidoTask = "other"
			return m
		},
		"handoff": func(m manifest) manifest {
			m.Infra.Paths = nil
			return m
		},
		"visibility": func(m manifest) manifest {
			m.InfraVisibility.MustKnow = nil
			return m
		},
		"direction": func(m manifest) manifest {
			m.DependencyDirection.TopDown = ""
			return m
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := verifyInfraBoundary(mutate(testManifest(t))); err == nil {
				t.Fatalf("verifyInfraBoundary accepted %s drift", name)
			}
		})
	}
}
