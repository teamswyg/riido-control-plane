package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/requirements"
	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/smokematrix"
)

func verifyContractMirror(root string, m manifest) error {
	for _, path := range []string{m.DSL, m.IR, m.OpenAPI, m.SmokeMatrix, m.GeneratedCore, m.GeneratedReact} {
		if err := verifyFile(pathutil.Resolve(root, path)); err != nil {
			return err
		}
	}
	ops, err := loadOpenAPIOperations(pathutil.Resolve(root, m.OpenAPI))
	if err != nil {
		return fmt.Errorf("load openapi operations: %w", err)
	}
	if ops.Counts != m.OperationCounts.withoutSmoke() {
		return fmt.Errorf("operation counts = %+v, want %+v", ops.Counts, m.OperationCounts)
	}
	smokePaths, smokeCount, err := smokematrix.LoadGeneratedPaths(pathutil.Resolve(root, m.SmokeMatrix))
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
	for _, path := range append(requirements.RequiredGeneratedPaths, m.RequiredGeneratedPaths...) {
		if _, ok := openAPI[path]; !ok {
			return fmt.Errorf("required generated path %q missing from openapi", path)
		}
		if _, ok := smoke[path]; !ok {
			return fmt.Errorf("required generated path %q missing from smoke matrix", path)
		}
	}
	return nil
}
