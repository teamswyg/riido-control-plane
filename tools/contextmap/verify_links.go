package main

import (
	"fmt"
	"os"
)

func verifyLinks(root string, links []link) error {
	if len(links) == 0 {
		return fmt.Errorf("ssot links are required")
	}
	for _, item := range links {
		if item.Name == "" || item.Path == "" {
			return fmt.Errorf("ssot link name and path are required")
		}
		if _, err := os.Stat(resolve(root, "docs/20-domain/"+item.Path)); err != nil {
			if _, err = os.Stat(resolve(root, item.Path)); err != nil {
				return fmt.Errorf("ssot link %q missing target %q: %w", item.Name, item.Path, err)
			}
		}
	}
	return nil
}
