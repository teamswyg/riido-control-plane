package main

import "testing"

func assertEvidenceCoverage(t *testing.T, got evidence) {
	t.Helper()
	assertGeneratedCoverage(t, got)
	if got.DirectSSOTCount != got.DirectLoopCount || len(got.DirectMissingLoop) != 0 {
		t.Fatalf("direct SSOT loop coverage drifted: %+v", got)
	}
	if got.StandaloneManifestCount != got.StandaloneManifestBindingCount ||
		len(got.StandaloneMissingBinding) != 0 {
		t.Fatalf("standalone manifest coverage drifted: %+v", got)
	}
	if got.SourceManifestCount != got.SourceManifestMetadataCount ||
		len(got.SourceMissingMetadata) != 0 {
		t.Fatalf("source manifest metadata coverage drifted: %+v", got)
	}
	if got.SourceManifestCount != got.SourceManifestBindingCount ||
		len(got.SourceMissingBinding) != 0 {
		t.Fatalf("source manifest coverage drifted: %+v", got)
	}
	if got.ContractArtifactCount != got.ContractArtifactBindingCount ||
		len(got.ContractMissingBinding) != 0 {
		t.Fatalf("contract artifact coverage drifted: %+v", got)
	}
	if got.ImportedManifestCount != got.ImportedManifestBindingCount ||
		len(got.ImportedMissingBinding) != 0 {
		t.Fatalf("imported manifest coverage drifted: %+v", got)
	}
	if got.OwnedManifestCount != got.OwnedManifestBindingCount ||
		len(got.OwnedMissingBinding) != 0 {
		t.Fatalf("owned manifest coverage drifted: %+v", got)
	}
	if got.ManifestInventoryCount != got.TrackedManifestCount ||
		len(got.UntrackedManifests) != 0 {
		t.Fatalf("manifest inventory coverage drifted: %+v", got)
	}
	if len(got.ManifestInventoryByGroup) == 0 {
		t.Fatalf("manifest inventory breakdown is missing: %+v", got)
	}
}

func assertGeneratedCoverage(t *testing.T, got evidence) {
	t.Helper()
	if got.GeneratedCount != got.GeneratedToolCount || len(got.GeneratedMissingWorkflow) != 0 {
		t.Fatalf("generated tool coverage drifted: %+v", got)
	}
	if got.GeneratedCount != got.GeneratedEvidenceWorkflowCount ||
		len(got.GeneratedMissingEvidenceWorkflow) != 0 {
		t.Fatalf("generated evidence workflow coverage drifted: %+v", got)
	}
	if got.GeneratedCount != got.GeneratedDeclaredWorkflowCount ||
		len(got.GeneratedMissingDeclaredWorkflow) != 0 {
		t.Fatalf("generated declared workflow coverage drifted: %+v", got)
	}
	if got.GeneratedCount != got.GeneratedArtifactBindingCount ||
		len(got.GeneratedMissingArtifactBinding) != 0 {
		t.Fatalf("generated artifact binding coverage drifted: %+v", got)
	}
}
