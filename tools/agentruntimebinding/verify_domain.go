package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/agentruntimebinding/requirements"
)

func verifyDomain(m manifest) error {
	for _, name := range requirements.RequiredFields {
		if !hasRequiredField(m.BindingFields, name) {
			return fmt.Errorf("missing required binding field %q", name)
		}
	}
	if hasRequiredField(m.BindingFields, "device_id") {
		return fmt.Errorf("device_id must stay optional")
	}
	for _, id := range requirements.RequiredRules {
		if !hasRule(m.BindingRules, id) {
			return fmt.Errorf("missing binding rule %q", id)
		}
	}
	for _, id := range requirements.RequiredDeviceRules {
		if !hasRule(m.DeviceRules, id) {
			return fmt.Errorf("missing device rule %q", id)
		}
	}
	return nil
}

func hasRequiredField(fields []field, name string) bool {
	for _, field := range fields {
		if field.Name == name && field.Required {
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
