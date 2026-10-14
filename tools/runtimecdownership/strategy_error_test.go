package main

import "testing"

func TestVerifyStrategiesRejectsDrift(t *testing.T) {
	t.Parallel()
	root := testRepoRoot(t)
	cases := map[string]func(manifest) manifest{
		"current-id": func(m manifest) manifest {
			m.Current.ID = "other"
			return m
		},
		"current-owner": func(m manifest) manifest {
			m.Current.CDOwner = "other"
			return m
		},
		"workflow": func(m manifest) manifest {
			m.Current.Workflow = "missing.yml"
			return m
		},
		"thumbnail-smoke": func(m manifest) manifest {
			m.Current.Allowed = []string{"resolving its digest instead of overwriting the tag"}
			return m
		},
		"digest": func(m manifest) manifest {
			m.Current.Allowed = []string{"profile thumbnail upload-intent smoke"}
			return m
		},
		"counts": func(m manifest) manifest {
			m.OptionalModes = nil
			return m
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := verifyStrategies(root, mutate(testManifest(t))); err == nil {
				t.Fatalf("verifyStrategies accepted %s drift", name)
			}
		})
	}
}
