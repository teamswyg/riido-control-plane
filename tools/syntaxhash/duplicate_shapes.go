package main

import (
	"fmt"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/syntaxhash/duplicates"
)

type (
	duplicatePolicy = duplicates.Policy
	duplicateRun    = duplicates.Run
)

func duplicateShapeEvidence(graph syntaxGraph, policy duplicatePolicy) duplicateRun {
	return duplicates.Build(duplicateTargets(graph.Targets), policy)
}

func duplicateTargets(targets []targetGraph) []duplicates.Target {
	out := make([]duplicates.Target, 0, len(targets))
	for _, target := range targets {
		files := make([]duplicates.File, 0, len(target.FileHashes))
		for _, file := range target.FileHashes {
			files = append(files, duplicates.File{Path: file.Path, ShapeHash: file.ShapeHash})
		}
		out = append(out, duplicates.Target{
			ID: target.ID, PackagePath: target.PackagePath, Files: files,
		})
	}
	return out
}

func renderDuplicateShapes(b *strings.Builder, graph syntaxGraph) {
	fmt.Fprintln(b, "## Duplicate AST Shapes")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- status: `%s`\n", graph.Duplicates.Status)
	fmt.Fprintf(b, "- group by: `%s`\n", graph.Duplicates.GroupBy)
	fmt.Fprintf(b, "- duplicate groups: `%d`\n", graph.Duplicates.GroupCount)
	fmt.Fprintf(b, "- duplicate files: `%d`\n", graph.Duplicates.FileCount)
	fmt.Fprintf(b, "- internal groups: `%d`\n", graph.Duplicates.InternalGroupCount)
	if len(graph.Duplicates.Groups) == 0 {
		return
	}
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Shape Hash | Files | Packages |")
	fmt.Fprintln(b, "| --- | ---: | --- |")
	for _, group := range graph.Duplicates.Groups {
		fmt.Fprintf(b, "| `%s` | %d | `%s` |\n",
			shortHash(group.ShapeHash), group.FileCount, strings.Join(group.Packages, "`, `"))
	}
}
