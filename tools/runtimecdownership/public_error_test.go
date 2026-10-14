package main

import "testing"

func TestVerifyPublicBoundaryRejectsDrift(t *testing.T) {
	t.Parallel()
	cases := map[string]func(manifest) manifest{
		"export-task": func(m manifest) manifest {
			m.PublicExport.RiidoTask = "other"
			return m
		},
		"operational-task": func(m manifest) manifest {
			m.PublicOperationalDetailMinimization.RiidoTask = "other"
			return m
		},
		"scan-scope": func(m manifest) manifest {
			m.PublicSurfaceScan.ScopePaths = nil
			return m
		},
		"scan-output": func(m manifest) manifest {
			m.PublicSurfaceScan.WorkflowForbiddenMechanism = nil
			return m
		},
		"config-task": func(m manifest) manifest {
			m.PublicConfigKeyMinimization.RiidoTask = "other"
			return m
		},
		"sensitive-flag": func(m manifest) manifest {
			m.PublicSensitiveSurfaceGuard.PublicKeyNamesAreSensitive = false
			return m
		},
		"canonical-missing": func(m manifest) manifest {
			m.PublicSensitiveSurfaceGuard.CanonicalCDKeyListPaths = nil
			return m
		},
		"canonical-doc": func(m manifest) manifest {
			m.PublicSensitiveSurfaceGuard.CanonicalCDKeyListPaths = []string{generatedDoc}
			return m
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, _, _, _, err := verifyPublicBoundary(mutate(testManifest(t))); err == nil {
				t.Fatalf("verifyPublicBoundary accepted %s drift", name)
			}
		})
	}
}
