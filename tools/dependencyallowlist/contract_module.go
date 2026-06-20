package main

import (
	"fmt"
	"strings"
)

func verifyAllowedModule(i int, module allowedModule, seen map[string]struct{}) error {
	if strings.TrimSpace(module.Path) == "" {
		return fmt.Errorf("allowed_direct_modules[%d].path is required", i)
	}
	if strings.TrimSpace(module.Layer) == "" || strings.TrimSpace(module.Owner) == "" || strings.TrimSpace(module.Reason) == "" {
		return fmt.Errorf("allowed_direct_modules[%d] must include layer, owner, and reason", i)
	}
	if _, ok := allowedLayers[module.Layer]; !ok {
		return fmt.Errorf("allowed_direct_modules[%d].layer %q is not in vocabulary: %s", i, module.Layer, formatAllowedLayers())
	}
	if !module.Approved {
		return fmt.Errorf("allowed_direct_modules[%d].approved must be true for %q", i, module.Path)
	}
	if _, ok := seen[module.Path]; ok {
		return fmt.Errorf("duplicate allowed direct module %q", module.Path)
	}
	seen[module.Path] = struct{}{}
	return nil
}
