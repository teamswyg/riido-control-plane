package riidoaiserver

import (
	"strings"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func progressLineReplacesPrevious(line AgentThreadProgressLine) bool {
	return strings.TrimSpace(line.MessageKey) == progressmessage.AssistantPartialKey
}
