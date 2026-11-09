package riidoaiserver

import (
	"context"
	"reflect"
)

func (s *PersistentAIAgentClientStore) loadDeviceBootstrap(ctx context.Context, id string, device deviceBootstrapAuth) (deviceBootstrapRecord, error) {
	repo, ok := s.snapshotStore.(deviceBootstrapRepository)
	if !ok {
		return deviceBootstrapRecord{}, errCLIInstallationUnavailable
	}
	record, found, err := repo.readDeviceBootstrap(ctx, id)
	if err != nil {
		return deviceBootstrapRecord{}, err
	}
	if !found {
		return deviceBootstrapRecord{}, errCLIInstallationInvalid
	}
	return s.currentDeviceBootstrap(ctx, record, &device)
}

func (s *PersistentAIAgentClientStore) currentDeviceBootstrap(ctx context.Context, record deviceBootstrapRecord, device *deviceBootstrapAuth) (deviceBootstrapRecord, error) {
	snapshot, found, err := s.snapshotStore.LoadAIAgentClientSnapshot(ctx)
	if err != nil {
		return deviceBootstrapRecord{}, err
	}
	expected := record.Before
	if record.State == "complete" {
		expected = record.After
	}
	current, exists := snapshotDeviceCredential(snapshot, expected.DeviceID)
	if !found || !exists || !reflect.DeepEqual(current, expected) {
		return deviceBootstrapRecord{}, errCLIInstallationConflict
	}
	if device != nil && (device.deviceID != expected.DeviceID || device.ownerPrincipalID != expected.OwnerPrincipalID || device.executionCoverageGeneration != expected.ExecutionCoverageGeneration || device.credentialGeneration != expected.CredentialGeneration) {
		return deviceBootstrapRecord{}, ErrAuthorizationForbidden
	}
	if err := s.snapshotStore.(deviceBootstrapRepository).verifyBootstrapState(ctx, record); err != nil {
		return deviceBootstrapRecord{}, err
	}
	return record, nil
}
