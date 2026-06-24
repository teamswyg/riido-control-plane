package main

import (
	"bytes"
	"os/exec"
	"strings"
)

func gitOutput(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	return cmd.Output()
}

func gitChangedFiles(root, baseRef string) (map[string]bool, error) {
	out, err := gitOutput(root, "diff", "--name-only", baseRef+"...HEAD", "--")
	if err != nil {
		return nil, err
	}
	files := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files[line] = true
		}
	}
	return files, nil
}

func gitManifest(root, baseRef, manifestPath string) (manifest, error) {
	var m manifest
	out, err := gitOutput(root, "show", baseRef+":"+manifestPath)
	if err != nil {
		return m, err
	}
	return decodeManifest(bytes.NewReader(out))
}
