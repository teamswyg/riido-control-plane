package main

import "fmt"

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
