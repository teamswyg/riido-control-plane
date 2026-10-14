package main

import "testing"

func TestVerifyRejectsRequiredDomainDrift(t *testing.T) {
	t.Parallel()
	cases := map[string]func(manifest) manifest{
		"identity": func(m manifest) manifest {
			m.ID = "other"
			return m
		},
		"title": func(m manifest) manifest {
			m.Title = ""
			return m
		},
		"owned": func(m manifest) manifest {
			m.OwnedContexts = m.OwnedContexts[1:]
			return m
		},
		"imported": func(m manifest) manifest {
			m.ImportedContexts = m.ImportedContexts[1:]
			return m
		},
		"external": func(m manifest) manifest {
			m.ExternalContexts = m.ExternalContexts[1:]
			return m
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := mutate(validTestManifest())
			root := writeTestRepo(t, m, true)
			if err := verify(root, m, true); err == nil {
				t.Fatalf("verify accepted %s drift", name)
			}
		})
	}
}

func TestVerifyRejectsMissingOwnedPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	items := []ownedContext{{ID: "ctx", OwnerPaths: []string{"missing/path"}}}
	if err := verifyOwnedPaths(root, items); err == nil {
		t.Fatal("verifyOwnedPaths accepted a missing owner path")
	}
}
