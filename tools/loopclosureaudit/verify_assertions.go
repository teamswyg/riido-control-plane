package main

import (
	"fmt"
	"strings"
)

func verifyAssertions(assertions []string) error {
	if len(assertions) == 0 {
		return fmt.Errorf("audit must declare executable assertions")
	}
	for _, assertion := range assertions {
		if strings.TrimSpace(assertion) == "" {
			return fmt.Errorf("audit assertion must not be empty")
		}
	}
	return nil
}
