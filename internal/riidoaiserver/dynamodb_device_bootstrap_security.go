package riidoaiserver

import (
	"context"
	"reflect"
)

func (s *DynamoDBAIAgentClientSnapshot) enrollmentSecurityActions(ctx context.Context, snapshot AIAgentClientSnapshot, credentials AWSCredentials) ([]dynamoDBTransactWriteItem, error) {
	m, bootstrap := ctx.Value(deviceBootstrapMutationKey{}).(deviceBootstrapMutation)
	if !bootstrap {
		action, err := s.enrollmentCoverageAction(ctx, snapshot, credentials)
		if err != nil || action == nil {
			return nil, err
		}
		return []dynamoDBTransactWriteItem{*action}, nil
	}
	if err := m.valid(); err != nil {
		return nil, err
	}
	native, ok := ctx.Value(deviceEnrollmentMutationContextKey{}).(deviceEnrollmentMutation)
	if !ok || !native.existed || native.deviceID != m.after.Before.DeviceID || !reflect.DeepEqual(native.before, m.after.Before) {
		return nil, errCLIInstallationConflict
	}
	proposed, ok := snapshotDeviceCredential(snapshot, native.deviceID)
	expected := m.after.Before
	if m.after.State == "complete" {
		expected = m.after.After
	}
	if !ok || !reflect.DeepEqual(proposed, expected) {
		return nil, errCLIInstallationConflict
	}
	policy, err := s.bootstrapPolicyAction(ctx, m, credentials)
	if err != nil {
		return nil, err
	}
	record, err := deviceBootstrapRecordAction(s.tableName, m)
	if err != nil {
		return nil, err
	}
	return []dynamoDBTransactWriteItem{policy, record}, nil
}

func (s *DynamoDBAIAgentClientSnapshot) readBootstrapLegacyRevision(ctx context.Context, device string) (uint64, error) {
	credentials, err := s.desktopPresenceCredentials(ctx)
	if err != nil {
		return 0, err
	}
	item, err := s.bootstrapGet(ctx, devicePolicyKey(device), credentials)
	if err != nil {
		return 0, err
	}
	// v2 or a prepared fence must never be reinterpreted as legacy.
	if len(item) != 0 && dynamoDBStringValue(item, "policy_schema") != devicePolicyRevisionSchema {
		return 0, errCLIInstallationConflict
	}
	return devicePolicyRevision(item, device)
}
