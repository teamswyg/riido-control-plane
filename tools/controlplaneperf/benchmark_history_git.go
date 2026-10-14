package main

import (
	"os/exec"
	"strings"
)

type gitInfo struct {
	Commit string
	Branch string
	Dirty  bool
}

func readGitInfo(root string) gitInfo {
	commit := gitOutput(root, "rev-parse", "HEAD")
	branch := gitOutput(root, "rev-parse", "--abbrev-ref", "HEAD")
	dirty := gitOutput(root, "status", "--porcelain") != ""
	return gitInfo{Commit: commit, Branch: branch, Dirty: dirty}
}

func gitOutput(root string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
