package main

import "fmt"

func verifyDomain(m manifest) error {
	for _, name := range requiredPorts {
		if !hasSurface(m.Ports, name) {
			return fmt.Errorf("missing port %q", name)
		}
	}
	for _, name := range requiredRecords {
		if !hasSurface(m.Records, name) {
			return fmt.Errorf("missing record %q", name)
		}
	}
	for _, id := range requiredReplayRules {
		if !hasRule(m.ReplayRules, id) {
			return fmt.Errorf("missing replay rule %q", id)
		}
	}
	for _, name := range requiredConstants {
		if !hasConstant(m.VersionConstants, name) {
			return fmt.Errorf("missing version constant %q", name)
		}
	}
	return nil
}

func hasSurface(surfaces []surface, name string) bool {
	for _, item := range surfaces {
		if item.Name == name && item.Role != "" {
			return true
		}
	}
	return false
}

func hasRule(rules []rule, id string) bool {
	for _, rule := range rules {
		if rule.ID == id && rule.Rule != "" {
			return true
		}
	}
	return false
}

func hasConstant(constants []constant, name string) bool {
	for _, item := range constants {
		if item.Name == name && item.Value != "" {
			return true
		}
	}
	return false
}
