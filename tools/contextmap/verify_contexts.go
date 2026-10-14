package main

import (
	"fmt"
	"os"
)

func verifyContexts(root string, m manifest) error {
	for _, id := range requiredOwnedContexts {
		if !hasOwned(m.OwnedContexts, id) {
			return fmt.Errorf("missing owned context %q", id)
		}
	}
	for _, id := range requiredImportedContexts {
		if !hasImported(m.ImportedContexts, id) {
			return fmt.Errorf("missing imported context %q", id)
		}
	}
	for _, id := range requiredExternalContexts {
		if !hasExternal(m.ExternalContexts, id) {
			return fmt.Errorf("missing external context %q", id)
		}
	}
	return verifyOwnedPaths(root, m.OwnedContexts)
}

func verifyOwnedPaths(root string, items []ownedContext) error {
	for _, item := range items {
		for _, path := range item.OwnerPaths {
			if _, err := os.Stat(resolve(root, path)); err != nil {
				return fmt.Errorf("owned context %q missing path %q: %w", item.ID, path, err)
			}
		}
	}
	return nil
}
