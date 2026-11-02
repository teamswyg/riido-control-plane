package riidoaiserver

import (
	"fmt"
	"unicode/utf8"
)

func validateAgentInstruction(field, value string) error {
	if utf8.RuneCountInString(value) > AgentInstructionMaxCharacters {
		return fmt.Errorf("%s must be %d characters or fewer", field, AgentInstructionMaxCharacters)
	}
	return nil
}
