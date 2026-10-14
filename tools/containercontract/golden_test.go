package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const containerContractEvidenceHash = "b2c8214cf7fb5c231ff895e39ef06d5eeb28ef34465ebf28752c35dae6d10d38"

func TestContainerContractBehaviorGolden(t *testing.T) {
	root := containerContractRepoRoot(t)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir repo root: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
	out := filepath.Join(t.TempDir(), "container-image-contract-evidence.json")
	contract := "packaging/containers/riido_ai_server_container.riido.json"
	if err := run([]string{"-contract", contract, "-evidence-out", out}, io.Discard); err != nil {
		t.Fatalf("run: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got := sha256Hex(body); got != containerContractEvidenceHash {
		t.Fatalf("evidence hash drifted: got %s want %s", got, containerContractEvidenceHash)
	}
}

func containerContractRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "../.."))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}
