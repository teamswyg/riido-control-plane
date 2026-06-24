package main

func claimRequiredImpactFiles(claim claimBinding) []string {
	surface := claimSurfaceFor(claim, nil)
	values := append([]string{}, surface.CodePaths...)
	values = append(values, surface.TestPaths...)
	return values
}
