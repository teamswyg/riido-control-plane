package main

func refreshWorkflowCommand(workflow string, args []string) string {
	command := "gh workflow run " + workflow + " --ref main"
	for _, arg := range args {
		command += " " + arg
	}
	return command
}
