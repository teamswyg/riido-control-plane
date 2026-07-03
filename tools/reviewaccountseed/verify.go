package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/pathutil"
)

func verify(root string, m manifest, results []caseEvidence, checkDoc bool) error {
	if err := verifySourceChecks(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyCaseNames(m.Cases, results); err != nil {
		return err
	}
	if checkDoc {
		return verifyDoc(root, m)
	}
	return nil
}

func verifySourceChecks(root string, checks []sourceCheck) error {
	for _, check := range checks {
		body, err := os.ReadFile(pathutil.Resolve(root, check.File))
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

func verifySeedTerms(root string, m manifest) error {
	body, err := os.ReadFile(pathutil.Resolve(root, m.SeedSSOT))
	if err != nil {
		return err
	}
	terms, err := forbiddenSeedTerms(m)
	if err != nil {
		return err
	}
	for _, term := range terms {
		if strings.Contains(string(body), term) {
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
