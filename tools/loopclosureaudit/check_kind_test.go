package main

import "testing"

func TestCheckKindRegistryCoversManifestKinds(t *testing.T) {
	m, _ := loadForTest(t)
	for _, req := range m.Requirements {
		for _, check := range req.Checks {
			if _, ok := checkKindByName(check.Kind); !ok {
				t.Fatalf("check kind %q is not registered", check.Kind)
			}
		}
	}
}

func TestCheckKindRegistryHasExecutableHandlers(t *testing.T) {
	seen := map[string]bool{}
	for _, spec := range checkKindSpecs {
		if spec.kind == "" || spec.key == nil ||
			spec.surface == nil || spec.verify == nil {
			t.Fatalf("incomplete check kind spec: %+v", spec)
		}
		if seen[spec.kind] {
			t.Fatalf("duplicate check kind spec: %s", spec.kind)
		}
		seen[spec.kind] = true
	}
}
