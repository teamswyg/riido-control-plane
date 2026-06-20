package main

import (
	"fmt"
	"os"
	"strings"
)

func verifySeedTerms(root string, m manifest) error {
	body, err := os.ReadFile(resolve(root, m.SeedSSOT))
	if err != nil {
		return err
	}
	text := string(body)
	terms, err := forbiddenSeedTerms(m)
	if err != nil {
		return err
	}
	for _, term := range terms {
		if strings.Contains(text, term) {
			return fmt.Errorf("seed SSOT contains forbidden term %s", term)
		}
	}
	return nil
}

func forbiddenSeedTerms(m manifest) ([]string, error) {
	terms := append([]string{}, m.ForbiddenSeedTerms...)
	for _, set := range m.ForbiddenSeedTermSets {
		switch set {
		case "aws_credential_env_names":
			terms = append(terms, "AWS_"+"ACCESS_KEY", "AWS_"+"SECRET")
		default:
			return nil, fmt.Errorf("unknown forbidden seed term set %s", set)
		}
	}
	return terms, nil
}
