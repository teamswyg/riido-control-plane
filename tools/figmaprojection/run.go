package main

import "fmt"

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
