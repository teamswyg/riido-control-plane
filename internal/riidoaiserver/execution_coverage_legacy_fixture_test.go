package riidoaiserver

import "context"

// Direct adapter fixtures stand in for Store's trusted legacy binding lookup.
// The production adapter still checks the uncovered policy in its claim CAS.
func legacyClaimFixtureContext() context.Context {
	return context.WithValue(context.Background(), executionAdmissionKey{}, &executionAdmission{deviceID: "fixture-legacy-device"})
}
