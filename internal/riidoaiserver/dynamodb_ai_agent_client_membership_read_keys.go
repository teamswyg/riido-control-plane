package riidoaiserver

import "context"

func membershipSnapshotReadKeys(ctx context.Context) []string {
	parts := append([]string{dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships}, onboardingOwnedPartNames...)
	parts = append(parts, dynamoDBAIAgentClientSnapshotPartDeviceConnections)
	parts = append(parts, dynamoDBAIAgentClientSnapshotPartDeviceCredentials, dynamoDBAIAgentClientSnapshotPartDevices)
	if fresh, _ := ctx.Value(taskAgentFreshSnapshotContextKey{}).(bool); fresh {
		parts = dynamoDBAIAgentClientSnapshotPartNames
	}
	keys := []string{dynamoDBAIAgentClientSnapshotSK}
	for _, part := range parts {
		keys = append(keys, dynamoDBAIAgentClientSnapshotPartSK(part))
	}
	return keys
}
