package main

import "fmt"

func verifyDomain(m manifest) error {
	for _, value := range []string{"admin"} {
		if !containsExact(m.Roles, value) {
			return fmt.Errorf("missing role %q", value)
		}
	}
	for _, value := range []string{"public", "private"} {
		if !containsExact(m.Visibilities, value) {
			return fmt.Errorf("missing visibility %q", value)
		}
	}
	for _, value := range []string{"read", "update", "delete"} {
		if !containsExact(m.Actions, value) {
			return fmt.Errorf("missing action %q", value)
		}
	}
	return verifyRulesAndRoutes(m)
}

func verifyRulesAndRoutes(m manifest) error {
	for _, id := range []string{"admin-all", "owner-all", "public-read", "private-deny", "public-mutation-deny"} {
		if !hasRule(m.VisibilityRules, id) {
			return fmt.Errorf("missing rule %q", id)
		}
	}
	for _, route := range requiredRoutes {
		if !containsExact(m.Routes, route) {
			return fmt.Errorf("missing route %q", route)
		}
	}
	return nil
}

func hasRule(rules []rule, id string) bool {
	for _, rule := range rules {
		if rule.ID == id && rule.Subject != "" && rule.Record != "" && rule.Reason != "" {
			return true
		}
	}
	return false
}
