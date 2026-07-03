package sourcechecks

import (
	"fmt"
	"os"
	"strings"
)

type Check struct {
	Name     string
	File     string
	Contains []string
}

func Verify(root string, checks []Check, resolve func(string, string) string) error {
	for _, check := range checks {
		body, err := os.ReadFile(resolve(root, check.File))
		if err != nil {
			return fmt.Errorf("read source check %s: %w", check.Name, err)
		}
		text := string(body)
		for _, needle := range check.Contains {
			if !strings.Contains(text, needle) {
				return fmt.Errorf("source check %s missing %q in %s", check.Name, needle, check.File)
			}
		}
	}
	return nil
}
