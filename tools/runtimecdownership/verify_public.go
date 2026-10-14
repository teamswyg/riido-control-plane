package main

import "fmt"

func verifyPublicBoundary(m manifest) (int, int, int, int, error) {
	if err := verifyPublicContracts(m); err != nil {
		return 0, 0, 0, 0, err
	}
	if err := verifyConfigGuard(m); err != nil {
		return 0, 0, 0, 0, err
	}
	forbidden := len(m.Redaction.MustNot) + len(m.PublicExport.ForbiddenPublicExports)
	forbidden += len(m.PublicSensitiveSurfaceGuard.ForbiddenPublicInformation)
	stableKeys := len(m.PublicConfigKeyMinimization.RequiredSecretKeys)
	stableKeys += len(m.PublicConfigKeyMinimization.RequiredVariableKeys)
	stableKeys += len(m.PublicConfigKeyMinimization.OptionalVariableKeys)
	return 4, 2, forbidden, stableKeys, nil
}

func verifyPublicContracts(m manifest) error {
	if m.PublicExport.RiidoTask != "RIID-4835" || m.PublicSurfaceScan.RiidoTask != "RIID-4836" {
		return fmt.Errorf("public export or scan work unit drifted")
	}
	if m.PublicOperationalDetailMinimization.RiidoTask != "RIID-4853" {
		return fmt.Errorf("public operational detail work unit drifted")
	}
	if !contains(m.PublicSurfaceScan.ScopePaths, generatedDoc) {
		return fmt.Errorf("public surface scan must cover generated reader doc")
	}
	if !contains(m.PublicSurfaceScan.WorkflowForbiddenMechanism, "GITHUB_OUTPUT") {
		return fmt.Errorf("public scan must forbid live value step outputs")
	}
	return nil
}

func verifyConfigGuard(m manifest) error {
	config := m.PublicConfigKeyMinimization
	guard := m.PublicSensitiveSurfaceGuard
	if config.RiidoTask != "RIID-4839" || guard.RiidoTask != "RIID-4842" {
		return fmt.Errorf("public config or sensitive guard work unit drifted")
	}
	if !guard.PublicKeyNamesAreSensitive {
		return fmt.Errorf("public key names must remain sensitivity-budgeted")
	}
	if !contains(guard.CanonicalCDKeyListPaths, defaultManifest) {
		return fmt.Errorf("CD key list must be canonical in machine-readable manifest")
	}
	if contains(guard.CanonicalCDKeyListPaths, generatedDoc) {
		return fmt.Errorf("generated reader doc must not be exact CD key-list source")
	}
	return nil
}
