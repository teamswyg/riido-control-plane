package main

import (
	"fmt"
	"os"
)

type options struct {
	Repo        string
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}

func run(opt options) error {
	root, err := repoRoot(opt.Repo)
	if err != nil {
		return err
	}
	var m manifest
	if err := readJSON(repoPath(root, opt.Manifest), &m); err != nil {
		return err
	}
	graph, err := buildGraph(root, m)
	if err != nil {
		return err
	}
	if err := verifyGraph(m, graph); err != nil {
		return err
	}
	doc := renderDoc(m, graph)
	if opt.WriteDoc {
		if err := os.WriteFile(repoPath(root, m.GeneratedDoc), []byte(doc), 0o644); err != nil {
			return err
		}
	}
	if opt.CheckDoc {
		body, err := os.ReadFile(repoPath(root, m.GeneratedDoc))
		if err != nil {
			return err
		}
		if string(body) != doc {
			return fmt.Errorf("generated doc stale: run go run ./tools/syntaxhash -write-doc")
		}
	}
	if opt.EvidenceOut != "" {
		return writeJSON(repoPath(root, opt.EvidenceOut), graph)
	}
	return nil
}
