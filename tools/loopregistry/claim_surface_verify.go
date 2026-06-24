package main

import "fmt"

func verifyClaimSurface(claim claimBinding) error {
	surface := claimSurfaceFor(claim, nil, nil)
	if len(surface.CodePaths)+len(surface.TestPaths) == 0 {
		return fmt.Errorf("claim %s must bind at least one code or test path", claim.ID)
	}
	if len(surface.TestPaths)+len(surface.Verifiers) == 0 {
		return fmt.Errorf("claim %s must bind test or verifier surface", claim.ID)
	}
	if len(surface.GeneratedDocs) == 0 {
		return fmt.Errorf("claim %s must bind generated doc surface", claim.ID)
	}
	return nil
}
