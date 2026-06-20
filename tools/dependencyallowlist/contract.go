package main

import (
	"errors"
	"fmt"
	"strings"
)

func verifyContract(c contract) error {
	if c.SchemaVersion != schemaVersion {
		return fmt.Errorf("schema_version = %q, want %q", c.SchemaVersion, schemaVersion)
	}
	if strings.TrimSpace(c.Service) == "" {
		return errors.New("service is required")
	}
	if strings.TrimSpace(c.Policy) == "" {
		return errors.New("policy is required")
	}
	return verifyAllowedModules(c.AllowedDirectModules)
}

func verifyAllowedModules(modules []allowedModule) error {
	seen := map[string]struct{}{}
	for i, module := range modules {
		if err := verifyAllowedModule(i, module, seen); err != nil {
			return err
		}
	}
	return nil
}
