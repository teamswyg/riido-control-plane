package main

type questionCommand struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	NextArtifact string `json:"next_artifact"`
	NextCommand  string `json:"next_command"`
}

func openQuestionCommands(questions []question) []questionCommand {
	commands := make([]questionCommand, 0, len(questions))
	for _, item := range questions {
		if item.Status != "open" {
			continue
		}
		commands = append(commands, questionCommand{
			ID:           item.ID,
			Owner:        item.Owner,
			NextArtifact: item.NextArtifact,
			NextCommand:  item.NextCommand,
		})
	}
	return commands
}
