package main

import "fmt"

func verifyStaticLists(m manifest) error {
	if err := requireStrings("runtime config", m.RuntimeConfigKeys, requiredRuntimeConfigKeys); err != nil {
		return err
	}
	if err := requireStrings("public field", m.PublicFields, requiredPublicFields); err != nil {
		return err
	}
	return requireStrings("deployment evidence", m.DeploymentEvidence, requiredDeploymentEvidence)
}

func requireStrings(kind string, got, required []string) error {
	seen := map[string]struct{}{}
	for _, value := range got {
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, ok := seen[value]; !ok {
			return fmt.Errorf("missing %s %q", kind, value)
		}
	}
	return nil
}
