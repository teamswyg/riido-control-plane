package main

import "testing"

func TestVerifyManifestRejectsDomainDrift(t *testing.T) {
	t.Parallel()
	cases := map[string]func(manifest) manifest{
		"schema": func(m manifest) manifest {
			m.SchemaVersion = "other"
			return m
		},
		"identity": func(m manifest) manifest {
			m.ID = ""
			return m
		},
		"evidence": func(m manifest) manifest {
			m.EvidenceArtifact = ""
			return m
		},
		"contract": func(m manifest) manifest {
			m.SharedContracts = m.SharedContracts[1:]
			return m
		},
		"workflow-count": func(m manifest) manifest {
			m.FocusedWorkflows = m.FocusedWorkflows[:1]
			return m
		},
		"boundary-count": func(m manifest) manifest {
			m.Boundaries = m.Boundaries[:1]
			return m
		},
		"loop": func(m manifest) manifest {
			m.Loop.Execute = ""
			return m
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := mutate(validTestManifest())
			root := writeTestRepo(t, m, true)
			if err := verifyManifest(root, m); err == nil {
				t.Fatalf("verifyManifest accepted %s drift", name)
			}
		})
	}
}
