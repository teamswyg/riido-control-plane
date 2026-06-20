package main

import (
	"fmt"
	"os"
	"strings"
)

func verifySourceChecks(root string, checks []sourceCheck) error {
	for _, check := range checks {
		body, err := os.ReadFile(resolve(root, check.File))
		if err != nil {
			return fmt.Errorf("read source check %s: %w", check.Name, err)
		}
		if err := verifyNeedles(check, string(body)); err != nil {
			return err
		}
	}
	return nil
}

func verifyNeedles(check sourceCheck, body string) error {
	for _, needle := range check.Contains {
		if !strings.Contains(body, needle) {
			return fmt.Errorf("source check %s missing %q in %s", check.Name, needle, check.File)
		}
	}
	return nil
}
