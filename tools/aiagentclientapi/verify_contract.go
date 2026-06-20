package main

import "fmt"

func verifyContractMirror(root string, m manifest) error {
	for _, path := range []string{m.DSL, m.IR, m.OpenAPI, m.SmokeMatrix, m.GeneratedCore, m.GeneratedReact} {
		if err := verifyFile(resolve(root, path)); err != nil {
			return err
		}
	}
	ops, err := loadOpenAPIOperations(resolve(root, m.OpenAPI))
	if err != nil {
		return fmt.Errorf("load openapi operations: %w", err)
	}
	if ops.Counts != m.OperationCounts.withoutSmoke() {
		return fmt.Errorf("operation counts = %+v, want %+v", ops.Counts, m.OperationCounts)
	}
	smokePaths, smokeCount, err := loadSmokeGeneratedPaths(resolve(root, m.SmokeMatrix))
	if err != nil {
		return fmt.Errorf("load smoke matrix: %w", err)
	}
	if smokeCount != m.OperationCounts.SmokeMatrix || len(smokePaths) != len(ops.Generated) {
		return fmt.Errorf("smoke matrix count mismatch")
	}
	return verifyRequiredPaths(m, ops.Generated, smokePaths)
}

func (c operationCounts) withoutSmoke() operationCounts {
	c.SmokeMatrix = 0
	return c
}

func verifyRequiredPaths(m manifest, openAPI, smoke map[string]struct{}) error {
	for _, path := range append(requiredGeneratedPaths, m.RequiredGeneratedPaths...) {
		if _, ok := openAPI[path]; !ok {
			return fmt.Errorf("required generated path %q missing from openapi", path)
		}
		if _, ok := smoke[path]; !ok {
			return fmt.Errorf("required generated path %q missing from smoke matrix", path)
		}
	}
	return nil
}
