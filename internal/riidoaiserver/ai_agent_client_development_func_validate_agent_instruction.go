package riidoaiserver

import (
	"errors"
	"unicode/utf8"
)

func validateAgentInstruction(value string) error {
	if utf8.RuneCountInString(value) > AgentInstructionMaxCharacters {
		return errors.New("instruction must be 1000 characters or fewer")
	}
	return nil
}
