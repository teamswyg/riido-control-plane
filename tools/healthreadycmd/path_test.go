package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRootRejectsTreeWithoutGoMod(t *testing.T) {
	_, err := findRepoRoot(filepath.Join(t.TempDir(), "nested"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want %v", err, os.ErrNotExist)
	}
}

func TestMustRunReturnsOnSuccessfulCommand(t *testing.T) {
	mustRun([]string{"-repo", "../.."})
}
