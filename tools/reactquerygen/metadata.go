package main

import (
	"errors"
	"fmt"
	"strings"
)

func validateClientMetadata(spec openAPISpec, ops []routeOperation) error {
	if len(spec.ClientModules) == 0 {
		return errors.New("OpenAPI x-riido-client-modules is required")
	}
	modules := map[string]struct{}{}
	for _, module := range spec.ClientModules {
		name := strings.TrimSpace(module.Module)
		if name == "" {
			return errors.New("x-riido-client-modules contains an empty module")
		}
		modules[name] = struct{}{}
	}
	cacheTags := map[string]routeOperation{}
	if err := validateCacheTags(ops, cacheTags); err != nil {
		return err
	}
	for _, op := range ops {
		if err := validateOperationMetadata(op, modules, cacheTags); err != nil {
			return err
		}
	}
	return nil
}

func validateCacheTags(ops []routeOperation, cacheTags map[string]routeOperation) error {
	for _, op := range ops {
		if !strings.EqualFold(op.Method, "GET") {
			continue
		}
		cacheTag := strings.TrimSpace(op.Op.Client.CacheTag)
		if cacheTag == "" {
			return fmt.Errorf("%s %s missing x-riido-client.cache_tag", strings.ToUpper(op.Method), op.Path)
		}
		if prev, exists := cacheTags[cacheTag]; exists {
			return fmt.Errorf("duplicate x-riido-client.cache_tag %q on %s and %s", cacheTag, prev.Op.OperationID, op.Op.OperationID)
		}
		cacheTags[cacheTag] = op
	}
	return nil
}

func validateOperationMetadata(op routeOperation, modules map[string]struct{}, cacheTags map[string]routeOperation) error {
	module := strings.TrimSpace(op.Op.Client.Module)
	if module == "" {
		return fmt.Errorf("%s %s missing x-riido-client.module", strings.ToUpper(op.Method), op.Path)
	}
	if _, ok := modules[module]; !ok {
		return fmt.Errorf("%s %s references unknown x-riido-client.module %q", strings.ToUpper(op.Method), op.Path, module)
	}
	return validateOperationMetadataMore(op, cacheTags)
}
