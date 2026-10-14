package controlplane

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestControlPlaneOwnerRegistryGenerationIsDeterministic(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test source")
	}
	directory := filepath.Dir(filename)
	temporary := t.TempDir()
	first := filepath.Join(temporary, "first.go")
	second := filepath.Join(temporary, "second.go")
	for _, output := range []string{first, second} {
		command := exec.Command("go", "run", "registry_gen.go",
			"-schema", "../../../contracts/nonwork17-owner-schema/owner-schema.graphqls", "-out", output)
		command.Dir = directory
		if raw, err := command.CombinedOutput(); err != nil {
			t.Fatalf("generate registry: %v: %s", err, raw)
		}
	}
	firstRaw, firstErr := os.ReadFile(first)
	secondRaw, secondErr := os.ReadFile(second)
	physical, physicalErr := os.ReadFile(filepath.Join(directory, "registry.generated.go"))
	if firstErr != nil || secondErr != nil || physicalErr != nil {
		t.Fatalf("read generated files: %v %v %v", firstErr, secondErr, physicalErr)
	}
	if !bytes.Equal(firstRaw, secondRaw) || !bytes.Equal(firstRaw, physical) {
		t.Fatal("two generated runs and committed registry must be byte-identical")
	}
}
