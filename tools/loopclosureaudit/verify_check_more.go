package main

import (
	"fmt"
	"strings"
)

func verifyWorkflowCheck(root string, c check) error {
	text, err := readText(repoPath(root, c.Path))
	if err != nil {
		return err
	}
	for _, needle := range c.Contains {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("workflow %s missing %q", c.Path, needle)
		}
	}
	return nil
}

func verifyGraphEdgeCheck(c check, idx indexes) error {
	edge := graphEdge{From: c.From, To: c.To, Relation: c.Relation}
	if _, ok := idx.edges[edge]; !ok {
		return fmt.Errorf("missing graph edge %+v", edge)
	}
	return nil
}

func verifyGraphChainCheck(c check, idx indexes) error {
	chain, ok := idx.chains[c.ID]
	if !ok {
		return fmt.Errorf("missing graph chain %s", c.ID)
	}
	for _, claim := range c.Claims {
		if !contains(chain.Claims, claim) {
			return fmt.Errorf("graph chain %s missing claim %s", c.ID, claim)
		}
	}
	return nil
}

func verifyPreCommitHookCheck(c check, idx indexes) error {
	if _, ok := idx.hooks[c.ID]; !ok {
		return fmt.Errorf("missing pre-commit hook %s", c.ID)
	}
	return nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
