package main

import (
	"fmt"
	"strings"
)

func validateOperationMetadataMore(op routeOperation, cacheTags map[string]routeOperation) error {
	if len(op.Op.Client.FacadePath) == 0 {
		return fmt.Errorf("%s %s missing x-riido-client.facade_path", strings.ToUpper(op.Method), op.Path)
	}
	for _, segment := range op.Op.Client.FacadePath {
		if strings.TrimSpace(segment) == "" {
			return fmt.Errorf("%s %s has empty x-riido-client.facade_path segment", strings.ToUpper(op.Method), op.Path)
		}
	}
	if generatedPath := strings.TrimSpace(op.Op.Client.GeneratedPath); generatedPath != "" {
		want := generatedPathFromClient(op.Op.Client)
		if generatedPath != want {
			return fmt.Errorf("%s %s has x-riido-client.generated_path %q, want %q", strings.ToUpper(op.Method), op.Path, generatedPath, want)
		}
	}
	for _, invalidates := range op.Op.Client.Invalidates {
		if _, ok := cacheTags[invalidates]; !ok {
			return fmt.Errorf("%s invalidates unknown cache tag %q", op.Op.OperationID, invalidates)
		}
	}
	return nil
}
