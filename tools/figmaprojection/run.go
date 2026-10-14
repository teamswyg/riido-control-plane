package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "figmaprojection:", err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("figmaprojection", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "repository root")
	projection := fs.String("projection", defaultProjection, "projection manifest")
	source := fs.String("source", defaultSource, "mirrored source manifest")
	writeDoc := fs.Bool("write-doc", false, "write generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "run the generated-client projection gate")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(*repo, *projection, *source, *evidenceOut, *writeDoc, *checkDoc)
}

func run(repo, projectionPath, sourcePath, evidenceOut string, writeDoc, check bool) error {
	root, err := findRepoRoot(repo)
	if err != nil {
		return err
	}
	p, err := loadJSONFile[projectionManifest](repoPath(root, projectionPath))
	if err != nil {
		return err
	}
	s, err := loadJSONFile[sourceManifest](repoPath(root, sourcePath))
	if err != nil {
		return err
	}
	if err := verifyProjection(p, s); err != nil {
		return err
	}
	doc := renderDoc(p, s)
	if writeDoc {
		if err := writeDocFile(root, doc); err != nil {
			return err
		}
	}
	if check {
		if err := checkGeneratedDoc(root, doc); err != nil {
			return err
		}
		if err := checkProjectionGate(root); err != nil {
			return err
		}
	}
	if evidenceOut != "" {
		if err := writeJSON(evidenceOut, newEvidence(p, s)); err != nil {
			return err
		}
	}
	fmt.Printf("figmaprojection: verified %d entries and %d generated annotations\n",
		len(p.Entries), s.AnnotationPolicy.LiveInspection.TotalAPIGeneratedAnnotations)
	return nil
}
