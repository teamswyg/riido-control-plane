package main

import "strings"

func loopUsesHarnessPromotion(root string, loop registeredLoop) (bool, error) {
	text, err := workflowText(root, loop.RefreshWorkflow)
	if err != nil {
		return false, err
	}
	return strings.Contains(text, "go run ./tools/harnesspromotion"), nil
}
