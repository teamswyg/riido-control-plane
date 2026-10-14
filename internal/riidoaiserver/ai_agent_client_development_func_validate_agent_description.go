package riidoaiserver

import (
	"errors"
	"unicode/utf8"
)

func validateAgentDescription(value string) error {
	if utf8.RuneCountInString(value) > AgentDescriptionMaxCharacters {
		return errors.New("description must be 160 characters or fewer")
	}
	return nil
}
