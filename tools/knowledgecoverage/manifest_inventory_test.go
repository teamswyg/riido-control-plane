package main

import (
	"strings"
	"testing"
)

func TestManifestInventoryRejectsUntrackedRiidoManifest(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "loose.riido.json", "{}")
	problems := validateManifestInventory(root, manifest{}, nil)
	if len(problems) == 0 || !strings.Contains(problems[0], "untracked executable manifest") {
		t.Fatalf("problems = %#v", problems)
	}
}

func TestManifestInventoryAcceptsScannedSiblingManifest(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "docs/known.md", "# Known\n")
	mustWrite(t, root, "docs/known.riido.json", "{}")
	docs := []docClass{{Path: "docs/known.md"}}
	problems := validateManifestInventory(root, manifest{}, docs)
	if len(problems) != 0 {
		t.Fatalf("problems = %#v", problems)
	}
}
