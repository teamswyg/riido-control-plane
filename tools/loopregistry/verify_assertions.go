package main

import (
	"fmt"
	"strings"
)

func verifyAssertions(m manifest) error {
	if !containsAssertion(m.Assertions, "all loops with evidence expiry") {
		return fmt.Errorf("loop registry assertions must describe all expiring loops")
	}
	for _, text := range append([]string{}, m.Assertions...) {
		if containsStaleExpiryScope(text) {
			return fmt.Errorf("loop registry assertion has stale expiry scope")
		}
	}
	if containsStaleExpiryScope(m.Loop.Execute) {
		return fmt.Errorf("loop registry execute loop has stale expiry scope")
	}
	return nil
}

func containsAssertion(assertions []string, want string) bool {
	for _, assertion := range assertions {
		if strings.Contains(assertion, want) {
			return true
		}
	}
	return false
}

func containsStaleExpiryScope(text string) bool {
	text = strings.ToLower(text)
	stale := []string{"at or below 24 hours", "each 24h loop", "within 24h"}
	for _, phrase := range stale {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}
