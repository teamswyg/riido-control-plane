package main

import (
	"fmt"
	"strings"
)

var secretNameFragments = []string{"TOKEN", "KEY", "CREDENTIALS"}

func verifyEnvParity(m manifest, sourceNames map[string]bool) error {
	seen := map[string]bool{}
	for _, entry := range m.Entries {
		if err := verifyEntry(entry); err != nil {
			return err
		}
		if seen[entry.Name] {
			return fmt.Errorf("duplicate manifest env %s", entry.Name)
		}
		seen[entry.Name] = true
		if !sourceNames[entry.Name] {
			return fmt.Errorf("manifest env %s is not read by %s", entry.Name, m.SourceDir)
		}
	}
	for _, name := range sortedKeys(sourceNames) {
		if !seen[name] {
			return fmt.Errorf("runtime env %s is missing from config manifest", name)
		}
	}
	return nil
}

func verifyEntry(entry entry) error {
	if entry.Name == "" || entry.Default == "" || entry.Owner == "" || entry.Meaning == "" {
		return fmt.Errorf("env entry must include name, default, owner, and meaning")
	}
	if entry.Sensitivity == "" {
		return fmt.Errorf("%s must declare sensitivity", entry.Name)
	}
	if hasSecretName(entry.Name) && entry.Sensitivity == "public" {
		return fmt.Errorf("%s has secret-like name but public sensitivity", entry.Name)
	}
	return nil
}

func hasSecretName(name string) bool {
	for _, fragment := range secretNameFragments {
		if strings.Contains(name, fragment) {
			return true
		}
	}
	return false
}
