package riidoaiserver

import (
	"strings"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func normalizeProgressLine(line AgentThreadProgressLine) (AgentThreadProgressLine, bool) {
	line.Message = strings.TrimSpace(line.Message)
	line.MessageKey = strings.TrimSpace(line.MessageKey)
	line.MessageArgs = copyStringMap(line.MessageArgs)
	if line.MessageCode <= 0 {
		if payload, ok := parseProgressMessagePayload(line.Message); ok {
			line.MessageCode = payload.Code
			line.MessageKey = firstNonEmptyProgressValue(line.MessageKey, payload.Key)
			line.MessageArgs = mergeProgressArgs(line.MessageArgs, payload.Args)
			if message := strings.TrimSpace(payload.Message); message != "" {
				line.Message = message
			}
		}
	}
	if rendered, key, ok := renderProgressMessage(line.MessageCode, line.MessageArgs); ok {
		line.Message = rendered
		if line.MessageKey == "" {
			line.MessageKey = key
		}
		line.MessageArgs = progressmessage.NormalizeArgsForCode(line.MessageCode, line.MessageArgs)
	}
	if line.Message == "" {
		return AgentThreadProgressLine{}, false
	}
	return line, true
}

func firstNonEmptyProgressValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
