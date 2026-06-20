package main

import "testing"

func TestGeneratedClientHandoffRejectsNonRiidoWorkBranch(t *testing.T) {
	out := t.TempDir()
	cfg := baseTestConfig(out)
	cfg.TargetBranch = "react-query-v0.0.99-0123456"
	err := run(cfg)
	assertErrorContains(t, err, "Riido work branchName")
}
