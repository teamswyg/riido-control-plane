package main

import "path/filepath"

type figmaProjectionPaths struct {
	manifest       string
	doc            string
	openAPI        string
	sourceCoverage string
}

func newFigmaProjectionPaths() figmaProjectionPaths {
	return figmaProjectionPaths{
		manifest:       filepath.Join("..", "..", "docs", "30-architecture", "figma-ai-agent-control-plane-projection.riido.json"),
		doc:            filepath.Join("..", "..", "docs", "30-architecture", "figma-ai-agent-control-plane-projection.md"),
		openAPI:        filepath.Join("..", "..", "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json"),
		sourceCoverage: filepath.Join("..", "..", "contracts", "ai-agent-client", "figma-ai-agent-coverage.riido.json"),
	}
}
