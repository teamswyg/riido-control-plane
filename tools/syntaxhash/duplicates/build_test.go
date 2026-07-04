package duplicates

import "testing"

func TestBuildGroupsSharedShapes(t *testing.T) {
	targets := []Target{
		{ID: "tool", PackagePath: "tools/a", Files: []File{{Path: "tools/a/run.go", ShapeHash: "shape"}}},
		{ID: "server", PackagePath: "internal/b", Files: []File{{Path: "internal/b/run.go", ShapeHash: "shape"}}},
	}
	run := Build(targets, Policy{
		Enabled: true, GroupBy: "ast_shape_hash",
		PackagePrefixes: []string{"tools/", "internal/"},
		MaxGroups:       10, Status: "evidence_only",
	})
	if run.GroupCount != 1 || run.FileCount != 2 || run.InternalGroupCount != 1 {
		t.Fatalf("duplicate evidence = %+v", run)
	}
	if got := run.Groups[0]; got.ShapeHash != "shape" || len(got.Packages) != 2 {
		t.Fatalf("duplicate group = %+v", got)
	}
}

func TestBuildHonorsPackagePrefixes(t *testing.T) {
	targets := []Target{
		{ID: "cmd", PackagePath: "cmd/app", Files: []File{{Path: "cmd/app/a.go", ShapeHash: "shape"}}},
		{ID: "tool", PackagePath: "tools/app", Files: []File{{Path: "tools/app/a.go", ShapeHash: "shape"}}},
	}
	run := Build(targets, Policy{Enabled: true, PackagePrefixes: []string{"tools/"}})
	if run.GroupCount != 0 || run.FileCount != 0 {
		t.Fatalf("unexpected duplicate outside prefix: %+v", run)
	}
}
