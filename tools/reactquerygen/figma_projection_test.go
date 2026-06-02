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
	verifySourceContractsManifestProvenance(t, sourceCoverage.StabilizedBy, manifest.SourceContractsManifest.StabilizedBy, docPath)
	if got, want := len(sourceCoverage.ExpectedPages), 3; got != want {
		t.Fatalf("source coverage expected_pages = %d, want %d", got, want)
	}
	if got, want := len(sourceCoverage.NonUITopLevelNodes), 12; got != want {
		t.Fatalf("source coverage non_ui_top_level_nodes = %d, want %d", got, want)
	}
	verifyMirroredNonUITopLevelInventory(t, sourceCoverage)
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
	verifyMirroredFigmaInspectionMethod(t, sourceCoverage.InspectionMethod, docText)
	verifyMirroredFigmaSupportingToolLimitations(t, manifest.MirroredSupportingToolLimitations, sourceCoverage.SupportingToolLimitations, manifest.SourceContractsManifest.StabilizedBy, docText)
	assertNoStaleFigmaNodeReference(t, filepath.Join("..", "..", "docs"), "164-50215")
	assertNoStaleFigmaNodeReference(t, filepath.Join("..", "..", "docs"), "164:50215")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "starter-agent")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "starter agent")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "future desktop or web clients")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "future desktop/web clients")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), "future client bootstrap")
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "contracts", "ai-agent-client"), "future client bootstrap")
	staleRuntimeHost := "desktop-api." + "riido.ai"
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "docs"), staleRuntimeHost)
	assertNoStaleControlPlanePhrase(t, filepath.Join("..", "..", "contracts", "ai-agent-client"), staleRuntimeHost)

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
	verifyRuntimeEndpointLabelProjection(t, sourceCoverage, docText)
	verifyFigmaAPIGeneratedAnnotations(t, sourceCoverage.APIGeneratedAnnotations, docText, generatedPaths, sourceGeneratedPaths, string(core), string(react))
	verifyFigmaAPIGeneratedAnnotationInventory(t, sourceCoverage.APIGeneratedAnnotationInventory, docText, generatedPaths, sourceGeneratedPaths, string(core), string(react))
	verifyLegacyNonUIAbsorptions(t, manifest.LegacyNonUIAbsorptions, sourceCoverage, docText, generatedPaths, string(core), string(react))
	verifyNonUIPlanningAbsorptions(t, manifest.NonUIPlanningAbsorptions, sourceCoverage, docText, generatedPaths, string(core), string(react))

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

func verifyLegacyNonUIAbsorptions(t *testing.T, absorptions []figmaProjectionLegacyAbsorption, sourceCoverage figmaSourceCoverageManifest, docText string, generatedPaths map[string]string, core, react string) {
	t.Helper()
	if got, want := len(absorptions), 7; got != want {
		t.Fatalf("legacy_non_ui_absorptions = %d, want %d", got, want)
	}
	sourcePrimary := map[string]figmaSourceCoverageEntry{}
	for _, entry := range sourceCoverage.Entries {
		sourcePrimary[entry.NodeID] = entry
	}
	sourceNonUI := map[string]figmaSourceCoverageEntry{}
	for _, entry := range sourceCoverage.NonUITopLevelNodes {
		sourceNonUI[entry.NodeID] = entry
	}
	seen := map[string]bool{}
	for _, absorption := range absorptions {
		if seen[absorption.NodeID] {
			t.Fatalf("duplicate legacy absorption node_id %q", absorption.NodeID)
		}
		seen[absorption.NodeID] = true
		if absorption.ProjectionStatus != "absorbed_by_current_ui_generated_client" {
			t.Fatalf("legacy absorption %q projection_status = %q", absorption.NodeID, absorption.ProjectionStatus)
		}
		if strings.TrimSpace(absorption.LocalScope) == "" {
			t.Fatalf("legacy absorption %q must explain local_scope", absorption.NodeID)
		}
		if len(absorption.RequiredGeneratedPaths) == 0 {
			t.Fatalf("legacy absorption %q must require generated paths", absorption.NodeID)
		}
		source, ok := sourceNonUI[absorption.NodeID]
		if !ok {
			t.Fatalf("legacy absorption %q is missing from mirrored contracts non_ui_top_level_nodes", absorption.NodeID)
		}
		if source.PageID != "0:1" || source.CoverageStatus != "covered" || source.EvidenceKind != "figma_legacy_wireframe_section" {
			t.Fatalf("legacy absorption %q source coverage is not a covered legacy Wireframe section: %+v", absorption.NodeID, source)
		}
		if absorption.SourceCoverageStatus != source.CoverageStatus {
			t.Fatalf("legacy absorption %q source_coverage_status = %q, source = %q", absorption.NodeID, absorption.SourceCoverageStatus, source.CoverageStatus)
		}
		if source.Name != absorption.Name {
			t.Fatalf("legacy absorption %q name = %q, source name = %q", absorption.NodeID, absorption.Name, source.Name)
		}
		if source.AbsorbedByTopLevelNodeID != absorption.AbsorbedByTopLevelNodeID {
			t.Fatalf("legacy absorption %q absorbed_by_top_level_node_id = %q, source = %q", absorption.NodeID, absorption.AbsorbedByTopLevelNodeID, source.AbsorbedByTopLevelNodeID)
		}
		absorbed, ok := sourcePrimary[absorption.AbsorbedByTopLevelNodeID]
		if !ok {
			t.Fatalf("legacy absorption %q references missing current UI source entry %q", absorption.NodeID, absorption.AbsorbedByTopLevelNodeID)
		}
		for _, path := range absorption.RequiredGeneratedPaths {
			if !hasString(source.GeneratedPaths, path) {
				t.Fatalf("legacy absorption %q requires generated path %q absent from mirrored non-UI source coverage", absorption.NodeID, path)
			}
			if !hasString(absorbed.GeneratedPaths, path) {
				t.Fatalf("legacy absorption %q requires generated path %q absent from absorbed current UI entry %q", absorption.NodeID, path, absorption.AbsorbedByTopLevelNodeID)
			}
			if route, ok := generatedPaths[path]; !ok {
				t.Fatalf("legacy absorption %q requires unknown generated path %q", absorption.NodeID, path)
			} else if strings.TrimSpace(route) == "" {
				t.Fatalf("legacy absorption %q generated path %q has empty route", absorption.NodeID, path)
			}
			requiredComment := "계약 generated path: `" + path + "`"
			if !strings.Contains(core, requiredComment) {
				t.Fatalf("core generated client missing %q for legacy absorption %q", requiredComment, absorption.NodeID)
			}
			if !strings.Contains(react, requiredComment) {
				t.Fatalf("react generated client missing %q for legacy absorption %q", requiredComment, absorption.NodeID)
			}
		}
		for _, needle := range []string{absorption.NodeID, absorption.Name, absorption.AbsorbedByTopLevelNodeID} {
			if !strings.Contains(docText, needle) {
				t.Fatalf("projection doc must mention legacy absorption %q", needle)
			}
		}
	}
}

func verifyNonUIPlanningAbsorptions(t *testing.T, absorptions []figmaProjectionPlanningAbsorption, sourceCoverage figmaSourceCoverageManifest, docText string, generatedPaths map[string]string, core, react string) {
	t.Helper()
	if got, want := len(absorptions), 1; got != want {
		t.Fatalf("non_ui_planning_absorptions = %d, want %d", got, want)
	}
	sourceNonUI := map[string]figmaSourceCoverageEntry{}
	for _, entry := range sourceCoverage.NonUITopLevelNodes {
		sourceNonUI[entry.NodeID] = entry
	}
	seen := map[string]bool{}
	for _, absorption := range absorptions {
		if seen[absorption.NodeID] {
			t.Fatalf("duplicate planning absorption node_id %q", absorption.NodeID)
		}
		seen[absorption.NodeID] = true
		if absorption.ProjectionStatus != "absorbed_by_onboarding_generated_client" {
			t.Fatalf("planning absorption %q projection_status = %q", absorption.NodeID, absorption.ProjectionStatus)
		}
		if strings.TrimSpace(absorption.LocalScope) == "" || strings.TrimSpace(absorption.NoNewEndpointReason) == "" {
			t.Fatalf("planning absorption %q must explain local scope and no_new_endpoint_reason: %+v", absorption.NodeID, absorption)
		}
		source, ok := sourceNonUI[absorption.NodeID]
		if !ok {
			t.Fatalf("planning absorption %q is missing from mirrored contracts non_ui_top_level_nodes", absorption.NodeID)
		}
		if source.PageID != "42:3014" || source.CoverageStatus != "covered" || source.EvidenceKind != "figma_planning_section" {
			t.Fatalf("planning absorption %q source coverage is not a covered planning section: %+v", absorption.NodeID, source)
		}
		if source.Name != absorption.Name {
			t.Fatalf("planning absorption %q name = %q, source name = %q", absorption.NodeID, absorption.Name, source.Name)
		}
		if absorption.SourceCoverageStatus != source.CoverageStatus {
			t.Fatalf("planning absorption %q source_coverage_status = %q, source = %q", absorption.NodeID, absorption.SourceCoverageStatus, source.CoverageStatus)
		}
		for _, path := range absorption.RequiredGeneratedPaths {
			if !hasString(source.GeneratedPaths, path) {
				t.Fatalf("planning absorption %q requires generated path %q absent from mirrored non-UI source coverage", absorption.NodeID, path)
			}
			if route, ok := generatedPaths[path]; !ok {
				t.Fatalf("planning absorption %q requires unknown generated path %q", absorption.NodeID, path)
			} else if strings.TrimSpace(route) == "" {
				t.Fatalf("planning absorption %q generated path %q has empty route", absorption.NodeID, path)
			}
			requiredComment := "계약 generated path: `" + path + "`"
			if !strings.Contains(core, requiredComment) {
				t.Fatalf("core generated client missing %q for planning absorption %q", requiredComment, absorption.NodeID)
			}
			if !strings.Contains(react, requiredComment) {
				t.Fatalf("react generated client missing %q for planning absorption %q", requiredComment, absorption.NodeID)
			}
		}
		for _, needle := range []string{absorption.NodeID, absorption.Name, "client-local", "workspace-less create"} {
			if !strings.Contains(docText, needle) {
				t.Fatalf("projection doc must mention planning absorption boundary %q", needle)
			}
		}
	}
}

func verifyRuntimeEndpointLabelProjection(t *testing.T, sourceCoverage figmaSourceCoverageManifest, docText string) {
	t.Helper()
	var evidenceFound bool
	for _, node := range sourceCoverage.VerifiedEvidenceNodes {
		if node.NodeID == "129:17930" {
			evidenceFound = true
			if !strings.Contains(strings.ToLower(node.Name), "endpoint") {
				t.Fatalf("source coverage runtime endpoint-looking evidence node must explain its role: %+v", node)
			}
		}
	}
	if !evidenceFound {
		t.Fatal("source coverage must register runtime settings endpoint-looking label node-id=129:17930")
	}
	var runtimeEntry figmaSourceCoverageEntry
	for _, entry := range sourceCoverage.Entries {
		if entry.NodeID == "162:23090" {
			runtimeEntry = entry
			break
		}
	}
	if runtimeEntry.NodeID == "" {
		t.Fatal("source coverage runtime settings entry 162:23090 is missing")
	}
	facts := strings.Join(runtimeEntry.CoveredFacts, "\n")
	for _, needle := range []string{
		"node-id=129:17930",
		"not a canonical base URL",
		"generated path",
		"live host export",
	} {
		if !strings.Contains(facts, needle) {
			t.Fatalf("source coverage runtime settings facts must classify endpoint-looking label with %q: %q", needle, facts)
		}
		if !strings.Contains(docText, needle) {
			t.Fatalf("projection doc must preserve endpoint-looking label boundary with %q", needle)
		}
	}
}

func verifyMirroredFigmaInspectionMethod(t *testing.T, method figmaCoverageInspectionMethod, docText string) {
	t.Helper()
	if method.ID != "figma-plugin-api-page-registry.v1" {
		t.Fatalf("source coverage inspection method id = %q", method.ID)
	}
	if method.PageRegistryExpression != "figma.root.children" {
		t.Fatalf("source coverage page registry expression = %q", method.PageRegistryExpression)
	}
	if method.TopLevelChildCountExpression != "await figma.setCurrentPageAsync(page); page.children.length" {
		t.Fatalf("source coverage top-level child count expression = %q", method.TopLevelChildCountExpression)
	}
	rule := strings.ToLower(method.Rule)
	for _, needle := range []string{"supporting evidence", "must not redefine page-level child counts", "lazy/unloaded"} {
		if !strings.Contains(rule, needle) {
			t.Fatalf("source coverage inspection rule must contain %q: %q", needle, method.Rule)
		}
	}
	for _, needle := range []string{"figma.root.children", "await figma.setCurrentPageAsync(page)", "page.children.length", "supporting evidence only", "lazy/unloaded"} {
		if !strings.Contains(docText, needle) {
			t.Fatalf("projection doc must mention mirrored inspection method with %q", needle)
		}
	}
}

func verifyMirroredFigmaSupportingToolLimitations(t *testing.T, projections []figmaProjectionSupportingToolLimitation, sourceLimitations []figmaSourceSupportingToolLimitation, stabilizedBy []string, docText string) {
	t.Helper()
	sourceByID := map[string]figmaSourceSupportingToolLimitation{}
	for _, limitation := range sourceLimitations {
		sourceByID[limitation.ID] = limitation
	}
	if len(projections) == 0 {
		t.Fatalf("mirrored_supporting_tool_limitations must record consumed source tooling limits")
	}
	var metadataProjection figmaProjectionSupportingToolLimitation
	var onboardingTimeoutProjection figmaProjectionSupportingToolLimitation
	for _, candidate := range projections {
		if candidate.SourceID == "figma-metadata-page-list-underreports-pages.v1" {
			metadataProjection = candidate
		}
		if candidate.SourceID == "figma-onboarding-page-load-timeout.v1" {
			onboardingTimeoutProjection = candidate
		}
	}
	if metadataProjection.SourceID == "" {
		t.Fatalf("mirrored_supporting_tool_limitations must include figma-metadata-page-list-underreports-pages.v1")
	}
	source, ok := sourceByID[metadataProjection.SourceID]
	if !ok {
		t.Fatalf("projection supporting limit %q is missing from mirrored contracts coverage", metadataProjection.SourceID)
	}
	if !strings.Contains(source.Tool, "get_metadata") || !strings.Contains(source.ObservedResult, "only page 129:5215 UI") {
		t.Fatalf("source supporting limit must preserve no-nodeId metadata under-report evidence: %+v", source)
	}
	if !hasString(stabilizedBy, "teamswyg/riido-contracts#52") {
		t.Fatalf("source_contracts_manifest.stabilized_by must include contracts #52 for mirrored metadata limitation: %+v", stabilizedBy)
	}
	if strings.TrimSpace(metadataProjection.LocalScope) == "" || !strings.Contains(metadataProjection.LocalScope, "must not collapse") {
		t.Fatalf("projection supporting limit must explain local scope: %+v", metadataProjection)
	}
	requiredPages := map[string]bool{"129:5215": false, "42:3014": false, "0:1": false}
	for _, pageID := range metadataProjection.RequiredAuthoritativePages {
		if _, ok := requiredPages[pageID]; ok {
			requiredPages[pageID] = true
		}
		if !hasString(source.AuthoritativeResult, pageID) {
			t.Fatalf("projection supporting limit page %q is absent from source authoritative_result: %+v", pageID, source.AuthoritativeResult)
		}
	}
	for pageID, seen := range requiredPages {
		if !seen {
			t.Fatalf("projection supporting limit is missing authoritative page %s", pageID)
		}
	}
	for _, forbidden := range []string{"remove expected_pages", "remove non_ui_top_level_inventory", "remove legacy_non_ui_absorptions"} {
		if !hasString(metadataProjection.ForbiddenProjectionEffects, forbidden) {
			t.Fatalf("projection supporting limit must forbid %q: %+v", forbidden, metadataProjection.ForbiddenProjectionEffects)
		}
	}
	for _, needle := range []string{"figma-metadata-page-list-underreports-pages.v1", "get_metadata", "teamswyg/riido-contracts#52", "`129:5215`", "`42:3014`", "`0:1`", "`expected_pages`", "`non_ui_top_level_inventory`", "`legacy_non_ui_absorptions`"} {
		if !strings.Contains(docText, needle) {
			t.Fatalf("projection doc must mention mirrored supporting tool limitation with %q", needle)
		}
	}
	if onboardingTimeoutProjection.SourceID == "" {
		t.Fatalf("mirrored_supporting_tool_limitations must include figma-onboarding-page-load-timeout.v1")
	}
	onboardingSource, ok := sourceByID[onboardingTimeoutProjection.SourceID]
	if !ok {
		t.Fatalf("projection supporting limit %q is missing from mirrored contracts coverage", onboardingTimeoutProjection.SourceID)
	}
	if !strings.Contains(onboardingSource.Tool, "42:3014") ||
		!strings.Contains(onboardingSource.ObservedResult, "timed out after 120s") ||
		!strings.Contains(onboardingSource.ObservedResult, "Wireframe - 온보딩") {
		t.Fatalf("source supporting limit must preserve onboarding load timeout evidence: %+v", onboardingSource)
	}
	if !hasString(onboardingSource.AuthoritativeResult, "42:3014") ||
		!hasString(onboardingSource.AuthoritativeResult, "child_count=83") ||
		!hasString(onboardingSource.AuthoritativeResult, "non_ui_top_level_inventory") {
		t.Fatalf("source onboarding timeout authoritative_result is incomplete: %+v", onboardingSource.AuthoritativeResult)
	}
	if strings.TrimSpace(onboardingTimeoutProjection.LocalScope) == "" ||
		!strings.Contains(onboardingTimeoutProjection.LocalScope, "must not treat") ||
		!strings.Contains(onboardingTimeoutProjection.LocalScope, "42:3014") {
		t.Fatalf("onboarding timeout projection must explain local scope: %+v", onboardingTimeoutProjection)
	}
	for _, pageID := range onboardingTimeoutProjection.RequiredAuthoritativePages {
		if !hasString(onboardingSource.AuthoritativeResult, pageID) {
			t.Fatalf("onboarding timeout projection page %q is absent from source authoritative_result: %+v", pageID, onboardingSource.AuthoritativeResult)
		}
	}
	for _, forbidden := range []string{"remove expected_pages", "remove non_ui_top_level_inventory", "remove onboarding generated paths", "mark onboarding generated paths unresolved"} {
		if !hasString(onboardingTimeoutProjection.ForbiddenProjectionEffects, forbidden) {
			t.Fatalf("onboarding timeout projection must forbid %q: %+v", forbidden, onboardingTimeoutProjection.ForbiddenProjectionEffects)
		}
	}
	for _, needle := range []string{"figma-onboarding-page-load-timeout.v1", "after 120s", "`Wireframe - 온보딩`", "`non_ui_top_level_inventory`", "mark onboarding generated paths unresolved"} {
		if !strings.Contains(docText, needle) {
			t.Fatalf("projection doc must mention onboarding page load timeout with %q", needle)
		}
	}
}

func verifySourceContractsManifestProvenance(t *testing.T, sourceStabilizedBy, projectionStabilizedBy []string, docPath string) {
	t.Helper()
	want := []string{
		"teamswyg/riido-contracts#38",
		"teamswyg/riido-contracts#39",
		"teamswyg/riido-contracts#45",
		"teamswyg/riido-contracts#46",
		"teamswyg/riido-contracts#51",
		"teamswyg/riido-contracts#52",
		"teamswyg/riido-contracts#54",
		"teamswyg/riido-contracts#55",
		"teamswyg/riido-contracts#56",
		"teamswyg/riido-contracts#57",
		"teamswyg/riido-contracts#58",
	}
	if len(sourceStabilizedBy) != len(want) {
		t.Fatalf("mirrored source coverage stabilized_by = %d entries, want %d: %+v", len(sourceStabilizedBy), len(want), sourceStabilizedBy)
	}
	for i := range want {
		if sourceStabilizedBy[i] != want[i] {
			t.Fatalf("mirrored source coverage stabilized_by[%d] = %q, want %q; full list = %+v", i, sourceStabilizedBy[i], want[i], sourceStabilizedBy)
		}
	}
	if len(projectionStabilizedBy) != len(sourceStabilizedBy) {
		t.Fatalf("source_contracts_manifest.stabilized_by = %d entries, mirrored source has %d: projection=%+v source=%+v", len(projectionStabilizedBy), len(sourceStabilizedBy), projectionStabilizedBy, sourceStabilizedBy)
	}
	for i := range sourceStabilizedBy {
		if projectionStabilizedBy[i] != sourceStabilizedBy[i] {
			t.Fatalf("source_contracts_manifest.stabilized_by[%d] = %q, mirrored source stabilized_by[%d] = %q; projection=%+v source=%+v", i, projectionStabilizedBy[i], i, sourceStabilizedBy[i], projectionStabilizedBy, sourceStabilizedBy)
		}
	}
	doc, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read projection doc for source provenance: %v", err)
	}
	docText := string(doc)
	for _, pr := range want {
		if !strings.Contains(docText, pr) {
			t.Fatalf("projection doc must mention full upstream contracts provenance %q", pr)
		}
	}
	if !strings.Contains(docText, "full upstream coverage provenance") ||
		!strings.Contains(docText, "limitation-local provenance") {
		t.Fatalf("projection doc must distinguish full upstream coverage provenance from limitation-local provenance")
	}
}

func verifyMirroredNonUITopLevelInventory(t *testing.T, sourceCoverage figmaSourceCoverageManifest) {
	t.Helper()
	pages := map[string]figmaSourceCoveragePage{}
	for _, page := range sourceCoverage.ExpectedPages {
		pages[page.NodeID] = page
	}
	if got, want := len(sourceCoverage.NonUITopLevelInventory), 2; got != want {
		t.Fatalf("source coverage non_ui_top_level_inventory pages = %d, want %d", got, want)
	}
	for _, inventory := range sourceCoverage.NonUITopLevelInventory {
		page, ok := pages[inventory.PageID]
		if !ok {
			t.Fatalf("non-UI inventory references unknown page %q", inventory.PageID)
		}
		if got, want := len(inventory.Nodes), page.ChildCount; got != want {
			t.Fatalf("non-UI inventory page %q nodes = %d, want loaded child_count %d", inventory.PageID, got, want)
		}
	}
	wireframe := pages["0:1"]
	if wireframe.ChildCount != 28 {
		t.Fatalf("Wireframe page loaded child_count = %d, want 28", wireframe.ChildCount)
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

func verifyFigmaAPIGeneratedAnnotations(t *testing.T, annotations []figmaSourceAPIGeneratedAnnotation, docText string, generatedPaths map[string]string, sourceGeneratedPaths map[string]map[string]bool, core, react string) {
	t.Helper()
	if got, want := len(annotations), 2; got != want {
		t.Fatalf("mirrored api_generated_annotations = %d, want %d", got, want)
	}
	seen := map[string]bool{}
	for _, annotation := range annotations {
		if seen[annotation.NodeID] {
			t.Fatalf("duplicate mirrored API Generated annotation %q", annotation.NodeID)
		}
		seen[annotation.NodeID] = true
		if annotation.CategoryID != "700:0" || annotation.CategoryLabel != "API Generated" {
			t.Fatalf("mirrored API Generated annotation %q category drifted: %+v", annotation.NodeID, annotation)
		}
		if !strings.HasPrefix(annotation.FigmaGeneratedPath, "riido.") {
			t.Fatalf("mirrored API Generated annotation %q must preserve Figma facade path: %q", annotation.NodeID, annotation.FigmaGeneratedPath)
		}
		canonical := strings.TrimPrefix(annotation.FigmaGeneratedPath, "riido.")
		if annotation.CanonicalGeneratedPath != canonical {
			t.Fatalf("mirrored API Generated annotation %q canonical path = %q, want %q", annotation.NodeID, annotation.CanonicalGeneratedPath, canonical)
		}
		if _, ok := generatedPaths[annotation.CanonicalGeneratedPath]; !ok {
			t.Fatalf("mirrored API Generated annotation %q references unknown generated path %q", annotation.NodeID, annotation.CanonicalGeneratedPath)
		}
		sourcePaths, ok := sourceGeneratedPaths[annotation.CoverageEntryNodeID]
		if !ok || !sourcePaths[annotation.CanonicalGeneratedPath] {
			t.Fatalf("mirrored API Generated annotation %q canonical path %q is not covered by source entry %q", annotation.NodeID, annotation.CanonicalGeneratedPath, annotation.CoverageEntryNodeID)
		}
		canonicalComment := "계약 generated path: `" + annotation.CanonicalGeneratedPath + "`"
		accessComment := "접근 예시: `" + annotation.FigmaGeneratedPath + "`"
		for _, generated := range []struct {
			name string
			body string
		}{
			{name: "core", body: core},
			{name: "react", body: react},
		} {
			if !strings.Contains(generated.body, canonicalComment) {
				t.Fatalf("%s generated client missing %q for mirrored API Generated annotation %q", generated.name, canonicalComment, annotation.NodeID)
			}
			if !strings.Contains(generated.body, accessComment) {
				t.Fatalf("%s generated client missing %q for mirrored API Generated annotation %q", generated.name, accessComment, annotation.NodeID)
			}
		}
		for _, needle := range []string{annotation.NodeID, annotation.FigmaGeneratedPath, annotation.CanonicalGeneratedPath, annotation.CategoryLabel} {
			if !strings.Contains(docText, needle) {
				t.Fatalf("projection doc must mention mirrored API Generated annotation %q", needle)
			}
		}
		if strings.Contains(annotation.FigmaLabel, "작업중") {
			if annotation.ResolutionStatus != "resolved_stale_handoff_copy" || !strings.Contains(annotation.Resolution, "stale") {
				t.Fatalf("mirrored API Generated annotation %q stale copy is not resolved: %+v", annotation.NodeID, annotation)
			}
			if !strings.Contains(docText, "상세내용은 작업중입니다") {
				t.Fatalf("projection doc must mention stale Figma handoff copy for annotation %q", annotation.NodeID)
			}
		}
	}
}

func verifyFigmaAPIGeneratedAnnotationInventory(t *testing.T, inventory []figmaSourceAPIGeneratedAnnotationGroup, docText string, generatedPaths map[string]string, sourceGeneratedPaths map[string]map[string]bool, core, react string) {
	t.Helper()
	if got, want := len(inventory), 19; got != want {
		t.Fatalf("mirrored api_generated_annotation_inventory = %d, want %d", got, want)
	}
	allowedKinds := map[string]bool{"Query": true, "Mutation": true, "SSE Stream": true}
	seenPath := map[string]bool{}
	totalAnnotations := 0
	for _, group := range inventory {
		if strings.TrimSpace(group.UIArea) == "" {
			t.Fatalf("mirrored API Generated inventory group has empty ui_area: %+v", group)
		}
		if group.CategoryID != "700:0" || group.CategoryLabel != "API Generated" {
			t.Fatalf("mirrored API Generated inventory group %q category drifted: %+v", group.FigmaGeneratedPath, group)
		}
		if !strings.HasPrefix(group.FigmaGeneratedPath, "riido.") {
			t.Fatalf("mirrored API Generated inventory group must preserve Figma facade path: %q", group.FigmaGeneratedPath)
		}
		canonical := strings.TrimPrefix(group.FigmaGeneratedPath, "riido.")
		if group.CanonicalGeneratedPath != canonical {
			t.Fatalf("mirrored API Generated inventory group %q canonical path = %q, want %q", group.FigmaGeneratedPath, group.CanonicalGeneratedPath, canonical)
		}
		if seenPath[group.CanonicalGeneratedPath] {
			t.Fatalf("duplicate mirrored API Generated inventory generated path %q", group.CanonicalGeneratedPath)
		}
		seenPath[group.CanonicalGeneratedPath] = true
		if _, ok := generatedPaths[group.CanonicalGeneratedPath]; !ok {
			t.Fatalf("mirrored API Generated inventory group references unknown generated path %q", group.CanonicalGeneratedPath)
		}
		if !allowedKinds[group.OperationKind] {
			t.Fatalf("mirrored API Generated inventory group %q operation_kind = %q", group.CanonicalGeneratedPath, group.OperationKind)
		}
		if strings.TrimSpace(group.Background) == "" {
			t.Fatalf("mirrored API Generated inventory group %q must preserve Korean background text", group.CanonicalGeneratedPath)
		}
		annotationCount := 0
		for _, source := range group.Sources {
			if strings.TrimSpace(source.PageID) == "" || strings.TrimSpace(source.TopLevelNodeID) == "" || strings.TrimSpace(source.CoverageEntryNodeID) == "" {
				t.Fatalf("mirrored API Generated inventory group %q has invalid source: %+v", group.CanonicalGeneratedPath, source)
			}
			sourcePaths, ok := sourceGeneratedPaths[source.CoverageEntryNodeID]
			if !ok || !sourcePaths[group.CanonicalGeneratedPath] {
				t.Fatalf("mirrored API Generated inventory group %q canonical path is not covered by source entry %q", group.CanonicalGeneratedPath, source.CoverageEntryNodeID)
			}
			if len(source.NodeIDs) == 0 {
				t.Fatalf("mirrored API Generated inventory group %q source %q must list node ids", group.CanonicalGeneratedPath, source.TopLevelNodeID)
			}
			annotationCount += len(source.NodeIDs)
		}
		if group.AnnotationCount != annotationCount {
			t.Fatalf("mirrored API Generated inventory group %q annotation_count = %d, want node count %d", group.CanonicalGeneratedPath, group.AnnotationCount, annotationCount)
		}
		totalAnnotations += annotationCount
		canonicalComment := "계약 generated path: `" + group.CanonicalGeneratedPath + "`"
		accessComment := "접근 예시: `" + group.FigmaGeneratedPath + "`"
		for _, generated := range []struct {
			name string
			body string
		}{
			{name: "core", body: core},
			{name: "react", body: react},
		} {
			if !strings.Contains(generated.body, canonicalComment) {
				t.Fatalf("%s generated client missing %q for mirrored inventory path %q", generated.name, canonicalComment, group.CanonicalGeneratedPath)
			}
			if !strings.Contains(generated.body, accessComment) {
				t.Fatalf("%s generated client missing %q for mirrored inventory path %q", generated.name, accessComment, group.CanonicalGeneratedPath)
			}
		}
		for _, needle := range []string{group.UIArea, group.FigmaGeneratedPath, group.CanonicalGeneratedPath, group.OperationKind, group.Background} {
			if !strings.Contains(docText, needle) {
				t.Fatalf("projection doc must mention mirrored API Generated inventory %q", needle)
			}
		}
	}
	if got, want := totalAnnotations, 59; got != want {
		t.Fatalf("mirrored API Generated inventory node annotations = %d, want %d", got, want)
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

func hasString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

type figmaProjectionManifest struct {
	SchemaVersion                     string                                    `json:"schema_version"`
	ID                                string                                    `json:"id"`
	RiidoTask                         string                                    `json:"riido_task"`
	SourceContractsManifest           figmaProjectionSourceManifest             `json:"source_contracts_manifest"`
	ProjectionPolicy                  figmaProjectionPolicy                     `json:"projection_policy"`
	MirroredSupportingToolLimitations []figmaProjectionSupportingToolLimitation `json:"mirrored_supporting_tool_limitations"`
	LegacyNonUIAbsorptions            []figmaProjectionLegacyAbsorption         `json:"legacy_non_ui_absorptions"`
	NonUIPlanningAbsorptions          []figmaProjectionPlanningAbsorption       `json:"non_ui_planning_absorptions"`
	Entries                           []figmaProjectionEntry                    `json:"entries"`
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

type figmaProjectionSupportingToolLimitation struct {
	SourceID                   string   `json:"source_id"`
	LocalScope                 string   `json:"local_scope"`
	RequiredAuthoritativePages []string `json:"required_authoritative_pages"`
	ForbiddenProjectionEffects []string `json:"forbidden_projection_effects"`
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

type figmaProjectionLegacyAbsorption struct {
	NodeID                   string   `json:"node_id"`
	Name                     string   `json:"name"`
	ProjectionStatus         string   `json:"projection_status"`
	SourceCoverageStatus     string   `json:"source_coverage_status"`
	AbsorbedByTopLevelNodeID string   `json:"absorbed_by_top_level_node_id"`
	LocalScope               string   `json:"local_scope"`
	RequiredGeneratedPaths   []string `json:"required_generated_paths"`
}

type figmaProjectionPlanningAbsorption struct {
	NodeID                 string   `json:"node_id"`
	Name                   string   `json:"name"`
	ProjectionStatus       string   `json:"projection_status"`
	SourceCoverageStatus   string   `json:"source_coverage_status"`
	LocalScope             string   `json:"local_scope"`
	RequiredGeneratedPaths []string `json:"required_generated_paths"`
	NoNewEndpointReason    string   `json:"no_new_endpoint_reason"`
}

type figmaSourceCoverageManifest struct {
	SchemaVersion                   string                                   `json:"schema_version"`
	ID                              string                                   `json:"id"`
	StabilizedBy                    []string                                 `json:"stabilized_by"`
	InspectionMethod                figmaCoverageInspectionMethod            `json:"inspection_method"`
	SupportingToolLimitations       []figmaSourceSupportingToolLimitation    `json:"supporting_tool_limitations"`
	ExpectedPages                   []figmaSourceCoveragePage                `json:"expected_pages"`
	NonUITopLevelInventory          []figmaSourceCoverageInventory           `json:"non_ui_top_level_inventory"`
	VerifiedEvidenceNodes           []figmaSourceCoverageNode                `json:"verified_evidence_nodes"`
	NonUITopLevelNodes              []figmaSourceCoverageEntry               `json:"non_ui_top_level_nodes"`
	APIGeneratedAnnotations         []figmaSourceAPIGeneratedAnnotation      `json:"api_generated_annotations"`
	APIGeneratedAnnotationInventory []figmaSourceAPIGeneratedAnnotationGroup `json:"api_generated_annotation_inventory"`
	Entries                         []figmaSourceCoverageEntry               `json:"entries"`
}

type figmaSourceCoveragePage struct {
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	ChildCount int    `json:"child_count"`
}

type figmaSourceCoverageInventory struct {
	PageID string                    `json:"page_id"`
	Nodes  []figmaSourceCoverageNode `json:"nodes"`
}

type figmaSourceCoverageNode struct {
	NodeID string `json:"node_id"`
	Name   string `json:"name"`
}

type figmaCoverageInspectionMethod struct {
	ID                           string `json:"id"`
	PageRegistryExpression       string `json:"page_registry_expression"`
	TopLevelChildCountExpression string `json:"top_level_child_count_expression"`
	Rule                         string `json:"rule"`
}

type figmaSourceSupportingToolLimitation struct {
	ID                  string   `json:"id"`
	Tool                string   `json:"tool"`
	ObservedResult      string   `json:"observed_result"`
	AuthoritativeResult []string `json:"authoritative_result"`
}

type figmaSourceCoverageEntry struct {
	PageID                   string   `json:"page_id,omitempty"`
	NodeID                   string   `json:"node_id"`
	Name                     string   `json:"name,omitempty"`
	CoverageStatus           string   `json:"coverage_status,omitempty"`
	EvidenceKind             string   `json:"evidence_kind,omitempty"`
	AbsorbedByTopLevelNodeID string   `json:"absorbed_by_top_level_node_id,omitempty"`
	GeneratedPaths           []string `json:"generated_paths,omitempty"`
	CoveredFacts             []string `json:"covered_facts,omitempty"`
}

type figmaSourceAPIGeneratedAnnotation struct {
	NodeID                 string `json:"node_id"`
	TopLevelNodeID         string `json:"top_level_node_id"`
	CoverageEntryNodeID    string `json:"coverage_entry_node_id"`
	CategoryID             string `json:"category_id"`
	CategoryLabel          string `json:"category_label"`
	FigmaLabel             string `json:"figma_label"`
	FigmaGeneratedPath     string `json:"figma_generated_path"`
	CanonicalGeneratedPath string `json:"canonical_generated_path"`
	ResolutionStatus       string `json:"resolution_status"`
	Resolution             string `json:"resolution"`
}

type figmaSourceAPIGeneratedAnnotationGroup struct {
	UIArea                 string                                    `json:"ui_area"`
	CategoryID             string                                    `json:"category_id"`
	CategoryLabel          string                                    `json:"category_label"`
	FigmaGeneratedPath     string                                    `json:"figma_generated_path"`
	CanonicalGeneratedPath string                                    `json:"canonical_generated_path"`
	OperationKind          string                                    `json:"operation_kind"`
	Background             string                                    `json:"background"`
	AnnotationCount        int                                       `json:"annotation_count"`
	Sources                []figmaSourceAPIGeneratedAnnotationSource `json:"sources"`
}

type figmaSourceAPIGeneratedAnnotationSource struct {
	PageID              string   `json:"page_id"`
	TopLevelNodeID      string   `json:"top_level_node_id"`
	CoverageEntryNodeID string   `json:"coverage_entry_node_id"`
	NodeIDs             []string `json:"node_ids"`
}
