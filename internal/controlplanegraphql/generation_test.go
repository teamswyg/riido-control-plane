package controlplanegraphql

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGeneratedReceiverIsCurrentAndDeterministicAcrossTwoRuns(t *testing.T) {
	directory := packageDirectory(t)
	outputs := []string{
		filepath.Join(directory, "generated", "generated.go"),
		filepath.Join(directory, "model", "models_gen.go"),
		filepath.Join(directory, "owner-schema.resolvers.go"),
	}
	committed := readOutputs(t, outputs)
	runGeneration(t, directory)
	first := readOutputs(t, outputs)
	runGeneration(t, directory)
	second := readOutputs(t, outputs)
	for index, path := range outputs {
		if !bytes.Equal(committed[index], first[index]) || !bytes.Equal(first[index], second[index]) {
			t.Fatalf("generated receiver output is stale or nondeterministic: %s", path)
		}
	}
}

func runGeneration(t *testing.T, directory string) {
	t.Helper()
	command := exec.Command("go", "generate", ".")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generate receiver: %v\n%s", err, output)
	}
}

func readOutputs(t *testing.T, paths []string) [][]byte {
	t.Helper()
	result := make([][]byte, len(paths))
	for index, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		result[index] = raw
	}
	return result
}

func packageDirectory(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve package directory")
	}
	return filepath.Dir(filename)
}
