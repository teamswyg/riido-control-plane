package riidoaiserver

import (
	"context"
	"reflect"
	"sort"
)

// Enrollment is the sole credential writer. Ordinary snapshot saves must not
// restore an older credential after a delayed or concurrent enrollment commit.
// The existing CURRENT hash CAS couples these reads and the one-device delta.
func mergeCommittedEnrollmentState(ctx context.Context, items []map[string]map[string]string, requested *AIAgentClientSnapshot) error {
	current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
	parts := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	var committed AIAgentClientSnapshot
	if dynamoDBAIAgentClientSnapshotItemIsLegacy(current) {
		var err error
		committed, err = decodeDesktopPresenceInline(current)
		if err != nil {
			return err
		}
	} else {
		for _, part := range []string{dynamoDBAIAgentClientSnapshotPartDeviceCredentials, dynamoDBAIAgentClientSnapshotPartDevices} {
			if err := decodeCommittedAIAgentClientPart(part, current, parts[part], &committed); err != nil {
				return err
			}
		}
	}
	mutation, enrolling := ctx.Value(deviceEnrollmentMutationContextKey{}).(deviceEnrollmentMutation)
	credentials := make(map[string]AIAgentClientDeviceCredentialSnapshot)
	for _, credential := range committed.DeviceCredentials {
		credentials[credential.DeviceID] = credential
	}
	if enrolling {
		old, exists := credentials[mutation.deviceID]
		if exists != mutation.existed || !reflect.DeepEqual(old, mutation.before) {
			return errAIAgentSnapshotConflict
		}
		proposed, exists := snapshotDeviceCredential(*requested, mutation.deviceID)
		if !exists {
			return errAIAgentSnapshotConflict
		}
		for deviceID, existing := range credentials {
			if deviceID != mutation.deviceID && proposed.MachineID != "" && existing.MachineID == proposed.MachineID {
				return errAIAgentSnapshotConflict
			}
		}
		credentials[mutation.deviceID] = proposed
	}
	requested.DeviceCredentials = make([]AIAgentClientDeviceCredentialSnapshot, 0, len(credentials))
	for _, credential := range credentials {
		requested.DeviceCredentials = append(requested.DeviceCredentials, credential)
	}
	sort.Slice(requested.DeviceCredentials, func(i, j int) bool {
		return requested.DeviceCredentials[i].DeviceID < requested.DeviceCredentials[j].DeviceID
	})
	mergeEnrollmentDevices(requested, committed, mutation.deviceID, enrolling)
	if sequence := parseDynamoDBAIAgentClientSnapshotInt(current, "next_device_credential_seq"); sequence > requested.NextDeviceCredentialSeq {
		requested.NextDeviceCredentialSeq = sequence
	}
	return nil
}
