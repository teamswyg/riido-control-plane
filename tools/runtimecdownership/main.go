package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "runtimecdownership:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("runtimecdownership", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "repository root")
	manifestPath := fs.String("manifest", defaultManifest, "runtime CD ownership manifest")
	writeDoc := fs.Bool("write-doc", false, "write the generated reader doc")
	checkDoc := fs.Bool("check-doc", false, "verify the generated reader doc")
	evidenceOut := fs.String("evidence-out", "", "optional evidence JSON output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return runWithOptions(*repo, *manifestPath, *writeDoc, *checkDoc, *evidenceOut)
}

func runWithOptions(repo, manifestPath string, writeDoc, checkDoc bool, evidenceOut string) error {
	root, err := findRepoRoot(repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(repoPath(root, manifestPath))
	if err != nil {
		return err
	}
	result, err := verifyAll(root, m)
	if err != nil {
		return err
	}
	doc := renderDoc(m, result)
	if writeDoc {
		if err := writeText(repoPath(root, generatedDoc), doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if checkDoc {
		if err := verifyDoc(root, doc); err != nil {
			return err
		}
	}
	if evidenceOut != "" {
		if err := writeJSON(evidenceOut, newEvidence(m, result)); err != nil {
			return err
		}
	}
	fmt.Printf("runtimecdownership: verified %d strategies, %d public policies\n", result.Strategies, result.PublicPolicies)
	return nil
}
