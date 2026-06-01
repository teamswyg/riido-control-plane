package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFigmaAIAgentControlPlaneProjectionManifest(t *testing.T) {
	manifestPath := filepath.Join("..", "..", "docs", "30-architecture", "figma-ai-agent-control-plane-projection.riido.json")
	docPath := filepath.Join("..", "..", "docs", "30-architecture", "figma-ai-agent-control-plane-projection.md")
	openAPIPath := filepath.Join("..", "..", "contracts", "ai-agent-client", "control-plane-ai-agent-client.openapi.json")
	sourceCoveragePath := filepath.Join("..", "..", "contracts", "ai-agent-client", "figma-ai-agent-coverage.riido.json")

	manifest := loadFigmaProjectionManifest(t, manifestPath)
	sourceCoverage := loadFigmaSourceCoverageManifest(t, sourceCoveragePath)
	if manifest.SchemaVersion != "riido-control-plane-figma-ai-agent-projection.v1" {
		t.Fatalf("schema_version = %q", manifest.SchemaVersion)
	}
	if manifest.ID != "figma-ai-agent-control-plane-generated-client-projection" {
		t.Fatalf("id = %q", manifest.ID)
	}
	if manifest.RiidoTask != "RIID-4810" {
		t.Fatalf("riido_task = %q", manifest.RiidoTask)
	}
	if manifest.SourceContractsManifest.Repo != "riido-contracts" ||
		manifest.SourceContractsManifest.Path != "docs/30-architecture/figma-ai-agent-coverage.riido.json" ||
		manifest.SourceContractsManifest.SchemaVersion != "riido-figma-ai-agent-coverage.v1" ||
		manifest.SourceContractsManifest.ID != "figma-v1-22-ai-agent-ui-coverage" {
		t.Fatalf("source contracts manifest = %+v", manifest.SourceContractsManifest)
	}
	if sourceCoverage.SchemaVersion != manifest.SourceContractsManifest.SchemaVersion ||
		sourceCoverage.ID != manifest.SourceContractsManifest.ID {
		t.Fatalf("source coverage mirror = %s/%s, want %s/%s", sourceCoverage.SchemaVersion, sourceCoverage.ID, manifest.SourceContractsManifest.SchemaVersion, manifest.SourceContractsManifest.ID)
	}
	if strings.TrimSpace(manifest.ProjectionPolicy.TopDown) == "" || strings.TrimSpace(manifest.ProjectionPolicy.BottomUp) == "" {
		t.Fatalf("projection policy must include both directions: %+v", manifest.ProjectionPolicy)
	}

	doc, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read projection doc: %v", err)
	}
	docText := string(doc)
	if !strings.Contains(docText, "does not redefine the Figma top-level UI coverage") {
		t.Fatalf("projection doc must name the downstream-only boundary")
	}
	assertNoStaleFigmaNodeReference(t, filepath.Join("..", "..", "docs"), "164-50215")
	assertNoStaleFigmaNodeReference(t, filepath.Join("..", "..", "docs"), "164:50215")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "starter-agent")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "starter agent")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "future desktop or web clients")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "future desktop/web clients")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "future client bootstrap")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "contracts", "ai-agent-client"), "future client bootstrap")

	spec, err := loadOpenAPI(openAPIPath)
	if err != nil {
		t.Fatalf("loadOpenAPI: %v", err)
	}
	core, err := generate(spec)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	react, err := generateReact(spec)
	if err != nil {
		t.Fatalf("generateReact: %v", err)
	}
	generatedPaths := generatedPathsByOperation(spec)
	generatedHaystack := generatedPathHaystack(spec, generatedPaths)
	sourceGeneratedPaths := sourceCoverageGeneratedPathsByNode(sourceCoverage)

	if got, want := len(manifest.Entries), 16; got != want {
		t.Fatalf("entries = %d, want %d", got, want)
	}
	seen := map[string]bool{}
	for _, entry := range manifest.Entries {
		if seen[entry.NodeID] {
			t.Fatalf("duplicate node_id %q", entry.NodeID)
		}
		seen[entry.NodeID] = true
		if !strings.Contains(docText, entry.NodeID) || !strings.Contains(docText, entry.Name) {
			t.Fatalf("projection doc must mention node %s %s", entry.NodeID, entry.Name)
		}
		verifyFigmaProjectionEntry(t, entry, sourceGeneratedPaths, generatedPaths, generatedHaystack, string(core), string(react))
	}
}

func assertNoStaleControlPlanePhrase(t *testing.T, root, phrase string) {
	t.Helper()
	needle := strings.ToLower(phrase)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".md", ".json":
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(strings.ToLower(string(data)), needle) {
				t.Fatalf("%s contains stale control-plane Figma wording %q; use onboarding fixture wording instead", path, phrase)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs for stale control-plane wording: %v", err)
	}
}

func assertNoStaleFigmaNodeReference(t *testing.T, root, staleNode string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".md", ".json":
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), staleNode) {
				t.Fatalf("%s still cites stale Figma node %s; use the contracts coverage manifest evidence nodes", path, staleNode)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs for stale Figma node references: %v", err)
	}
}

func loadFigmaProjectionManifest(t *testing.T, path string) figmaProjectionManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read projection manifest: %v", err)
	}
	var manifest figmaProjectionManifest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&manifest); err != nil {
		t.Fatalf("decode projection manifest: %v", err)
	}
	return manifest
}

func loadFigmaSourceCoverageManifest(t *testing.T, path string) figmaSourceCoverageManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read source coverage manifest: %v", err)
	}
	var manifest figmaSourceCoverageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode source coverage manifest: %v", err)
	}
	return manifest
}

func sourceCoverageGeneratedPathsByNode(manifest figmaSourceCoverageManifest) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, entry := range manifest.Entries {
		if _, ok := out[entry.NodeID]; !ok {
			out[entry.NodeID] = map[string]bool{}
		}
		for _, generatedPath := range entry.GeneratedPaths {
			out[entry.NodeID][generatedPath] = true
		}
	}
	return out
}

func verifyFigmaProjectionEntry(t *testing.T, entry figmaProjectionEntry, sourceGeneratedPaths map[string]map[string]bool, generatedPaths map[string]string, generatedHaystack, core, react string) {
	t.Helper()
	if strings.TrimSpace(entry.NodeID) == "" || strings.TrimSpace(entry.Name) == "" {
		t.Fatalf("entry has empty node id or name: %+v", entry)
	}
	if strings.TrimSpace(entry.SourceCoverageStatus) == "" {
		t.Fatalf("entry %q source_coverage_status is required", entry.NodeID)
	}
	switch entry.ProjectionStatus {
	case "generated_client_covered":
		if strings.TrimSpace(entry.LocalScope) == "" {
			t.Fatalf("entry %q local_scope is required", entry.NodeID)
		}
		if len(entry.RequiredGeneratedPaths) == 0 {
			t.Fatalf("entry %q must require generated paths", entry.NodeID)
		}
		sourcePaths, ok := sourceGeneratedPaths[entry.NodeID]
		if !ok {
			t.Fatalf("entry %q is missing from mirrored contracts Figma coverage", entry.NodeID)
		}
		for _, path := range entry.RequiredGeneratedPaths {
			if !sourcePaths[path] {
				t.Fatalf("entry %q requires generated path %q that is absent from mirrored contracts Figma coverage", entry.NodeID, path)
			}
			if route, ok := generatedPaths[path]; !ok {
				t.Fatalf("entry %q requires unknown generated path %q", entry.NodeID, path)
			} else if strings.TrimSpace(route) == "" {
				t.Fatalf("entry %q generated path %q has empty route", entry.NodeID, path)
			}
			requiredComment := "계약 generated path: `" + path + "`"
			if !strings.Contains(core, requiredComment) {
				t.Fatalf("core generated client missing %q for entry %q", requiredComment, entry.NodeID)
			}
			if !strings.Contains(react, requiredComment) {
				t.Fatalf("react generated client missing %q for entry %q", requiredComment, entry.NodeID)
			}
		}
	case "client_route_no_endpoint", "product_surface_no_endpoint", "planning_evidence", "non_decision_asset":
		if strings.TrimSpace(entry.NoEndpointReason) == "" {
			t.Fatalf("entry %q must explain why no endpoint is generated", entry.NodeID)
		}
		if len(entry.RequiredGeneratedPaths) != 0 {
			t.Fatalf("entry %q must not require generated paths for status %s", entry.NodeID, entry.ProjectionStatus)
		}
	default:
		t.Fatalf("entry %q has unknown projection_status %q", entry.NodeID, entry.ProjectionStatus)
	}
	for _, fragment := range entry.ForbiddenGeneratedPathFragments {
		if strings.Contains(generatedHaystack, strings.ToLower(fragment)) {
			t.Fatalf("entry %q forbids generated path fragment %q, but generated surface contains it", entry.NodeID, fragment)
		}
	}
}

func generatedPathsByOperation(spec openAPISpec) map[string]string {
	out := map[string]string{}
	for path, methods := range spec.Paths {
		for method, operation := range methods {
			generatedPath := operation.Client.GeneratedPath
			if strings.TrimSpace(generatedPath) == "" {
				generatedPath = generatedPathFromClient(operation.Client)
			}
			out[generatedPath] = strings.ToUpper(method) + " " + path + " " + operation.OperationID
		}
	}
	return out
}

func generatedPathHaystack(spec openAPISpec, generatedPaths map[string]string) string {
	parts := make([]string, 0, len(generatedPaths)+len(spec.Paths))
	for generatedPath, route := range generatedPaths {
		parts = append(parts, generatedPath+" "+route)
	}
	sort.Strings(parts)
	return strings.ToLower(strings.Join(parts, "\n"))
}

type figmaProjectionManifest struct {
	SchemaVersion           string                        `json:"schema_version"`
	ID                      string                        `json:"id"`
	RiidoTask               string                        `json:"riido_task"`
	SourceContractsManifest figmaProjectionSourceManifest `json:"source_contracts_manifest"`
	ProjectionPolicy        figmaProjectionPolicy         `json:"projection_policy"`
	Entries                 []figmaProjectionEntry        `json:"entries"`
}

type figmaProjectionSourceManifest struct {
	Repo          string   `json:"repo"`
	Path          string   `json:"path"`
	SchemaVersion string   `json:"schema_version"`
	ID            string   `json:"id"`
	StabilizedBy  []string `json:"stabilized_by"`
}

type figmaProjectionPolicy struct {
	Summary  string `json:"summary"`
	TopDown  string `json:"top_down"`
	BottomUp string `json:"bottom_up"`
}

type figmaProjectionEntry struct {
	NodeID                          string   `json:"node_id"`
	Name                            string   `json:"name"`
	ProjectionStatus                string   `json:"projection_status"`
	SourceCoverageStatus            string   `json:"source_coverage_status"`
	LocalScope                      string   `json:"local_scope,omitempty"`
	RequiredGeneratedPaths          []string `json:"required_generated_paths,omitempty"`
	ForbiddenGeneratedPathFragments []string `json:"forbidden_generated_path_fragments,omitempty"`
	NoEndpointReason                string   `json:"no_endpoint_reason,omitempty"`
}

type figmaSourceCoverageManifest struct {
	SchemaVersion string                     `json:"schema_version"`
	ID            string                     `json:"id"`
	Entries       []figmaSourceCoverageEntry `json:"entries"`
}

type figmaSourceCoverageEntry struct {
	NodeID         string   `json:"node_id"`
	GeneratedPaths []string `json:"generated_paths,omitempty"`
}
