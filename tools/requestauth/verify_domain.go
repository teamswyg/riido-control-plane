package main

import "fmt"

func verifyDomain(m manifest) error {
	for _, name := range requiredSurfaces {
		if !hasSurface(m.Surfaces, name) {
			return fmt.Errorf("missing surface %q", name)
		}
	}
	if err := requireStrings("resource", m.Resources, requiredResources); err != nil {
		return err
	}
	if err := requireStrings("contract version", m.ExternalContractVersions, requiredContractVersions); err != nil {
		return err
	}
	if err := requireStrings("runtime config", m.RuntimeConfigKeys, requiredRuntimeConfigKeys); err != nil {
		return err
	}
	for _, id := range requiredTransports {
		if !hasTransport(m.TokenTransports, id) {
			return fmt.Errorf("missing token transport %q", id)
		}
	}
	return verifyRuleGroups(m.RuleGroups)
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

func verifyRuleGroups(groups []ruleGroup) error {
	index := map[string][]string{}
	for _, group := range groups {
		index[group.ID] = group.Rules
	}
	for group, required := range requiredRuleGroups {
		if err := requireStrings("rule "+group, index[group], required); err != nil {
			return err
		}
	}
	return nil
}
