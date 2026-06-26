package main

func refreshWorkflowCommand(workflow string) string {
	return "gh workflow run " + workflow + " --ref main"
}
