package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

func gitOutput(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = gitEnvWithoutOuterRepo(os.Environ())
	return cmd.Output()
}

func gitEnvWithoutOuterRepo(env []string) []string {
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		if strings.HasPrefix(entry, "GIT_") {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func gitChangedFiles(root, baseRef string) (map[string]bool, error) {
	files := map[string]bool{}
	for _, args := range [][]string{
		{"diff", "--name-only", baseRef + "...HEAD", "--"},
		{"diff", "--name-only", "--cached", "--"},
		{"diff", "--name-only", "--"},
		{"ls-files", "--others", "--exclude-standard"},
	} {
		out, err := gitOutput(root, args...)
		if err != nil {
			return nil, err
		}
		addChangedFiles(files, out)
	}
	return files, nil
}

func addChangedFiles(files map[string]bool, out []byte) {
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files[line] = true
		}
	}
}

func gitManifest(root, baseRef, manifestPath string) (manifest, error) {
	var m manifest
	out, err := gitOutput(root, "show", baseRef+":"+manifestPath)
	if err != nil {
		return m, err
	}
	return m, json.NewDecoder(bytes.NewReader(out)).Decode(&m)
}
