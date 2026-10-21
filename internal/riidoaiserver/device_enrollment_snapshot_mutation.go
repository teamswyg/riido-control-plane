package riidoaiserver

import "context"

type deviceEnrollmentMutationContextKey struct{}

type deviceEnrollmentMutation struct {
	deviceID string
	before   AIAgentClientDeviceCredentialSnapshot
	existed  bool
}

func withDeviceEnrollmentMutation(ctx context.Context, before AIAgentClientSnapshot, deviceID string) context.Context {
	credential, existed := snapshotDeviceCredential(before, deviceID)
	return context.WithValue(ctx, deviceEnrollmentMutationContextKey{}, deviceEnrollmentMutation{deviceID: deviceID, before: credential, existed: existed})
}
