package main

import "testing"

type figmaGeneratedClientSurface struct {
	GeneratedPaths       map[string]string
	GeneratedHaystack    string
	SourceGeneratedPaths map[string]map[string]bool
	Core                 string
	React                string
}

func loadFigmaGeneratedClientSurface(t *testing.T, spec openAPISpec) figmaGeneratedClientSurface {
	t.Helper()
	core, err := generate(spec)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	react, err := generateReact(spec)
	if err != nil {
		t.Fatalf("generateReact: %v", err)
	}
	generatedPaths := generatedPathsByOperation(spec)
	return figmaGeneratedClientSurface{
		GeneratedPaths:    generatedPaths,
		GeneratedHaystack: generatedPathHaystack(spec, generatedPaths),
		Core:              string(core),
		React:             string(react),
	}
}

func (surface figmaGeneratedClientSurface) WithSourceCoverage(sourceCoverage figmaSourceCoverageManifest) figmaGeneratedClientSurface {
	surface.SourceGeneratedPaths = sourceCoverageGeneratedPathsByNode(sourceCoverage)
	return surface
}
