package main

import "testing"

func TestManifestLoopStatusAcceptsLoopSource(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "source.riido.json", manifestLoopFixture())
	mustWrite(t, root, "target.riido.json", `{"loop_source":"source.riido.json"}`)
	got := manifestLoopStatus(root, "target.riido.json", nil)
	if got != "delegated" {
		t.Fatalf("status = %q", got)
	}
}

func TestScanManifestLoopsAcceptsOwnerDelegation(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "docs/owner.riido.json", manifestLoopFixture())
	mustWrite(t, root, "contracts/artifact.riido.json", `{}`)
	m := manifest{
		ContractArtifacts: []contractArtifact{{
			Path: "contracts/artifact.riido.json", OwnerManifest: "docs/owner.riido.json",
		}},
	}
	got := scanManifestLoops(root, m)
	if got.Missing != 0 || got.Delegated != 1 || got.Direct != 1 {
		t.Fatalf("loops = %#v", got)
	}
}

func manifestLoopFixture() string {
	return `{"loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"}}`
}
