package main

import "strings"

var deprecatedActionRefs = []string{
	"actions/checkout@v4",
	"actions/setup-go@v5",
	"actions/upload-artifact@v4",
	"aws-actions/configure-aws-credentials@v4",
}

func deprecatedWorkflowActions(text string) []string {
	refs := []string{}
	for _, ref := range deprecatedActionRefs {
		if strings.Contains(text, ref) {
			refs = append(refs, ref)
		}
	}
	return refs
}
