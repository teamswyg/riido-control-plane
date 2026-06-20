package main

import "fmt"

func verifyDomain(m manifest) error {
	for _, name := range requiredSurfaces {
		if !hasSurface(m.Surfaces, name) {
			return fmt.Errorf("missing surface %q", name)
		}
	}
	for _, value := range requiredRoutingStatuses {
		if !hasValue(m.RoutingStatuses, value) {
			return fmt.Errorf("missing routing status %q", value)
		}
	}
	for _, value := range requiredDistributionChannels {
		if !hasValue(m.DistributionChannels, value) {
			return fmt.Errorf("missing distribution channel %q", value)
		}
	}
	if err := verifyRuleSet("validation", m.ValidationRules, requiredValidationRules); err != nil {
		return err
	}
	return verifyRuleSet("routing", m.RoutingRules, requiredRoutingRules)
}

func verifyRuleSet(kind string, got []rule, required []string) error {
	for _, id := range required {
		if !hasRule(got, id) {
			return fmt.Errorf("missing %s rule %q", kind, id)
		}
	}
	return nil
}
