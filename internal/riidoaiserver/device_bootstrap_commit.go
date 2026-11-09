package riidoaiserver

import (
	"context"
	"reflect"
	"time"
)

func (s *PersistentAIAgentClientStore) commitDeviceBootstrap(ctx context.Context, mutation deviceBootstrapMutation) (deviceBootstrapRecord, error) {
	var empty deviceBootstrapRecord
	if err := mutation.valid(); err != nil {
		return empty, err
	}
	if err := s.reloadSnapshot(ctx); err != nil {
		return empty, err
	}
	trial, before, err := s.enrollmentTrial()
	if err != nil {
		return empty, err
	}
	r := mutation.after
	current, found := snapshotDeviceCredential(before, r.Before.DeviceID)
	if !found || !reflect.DeepEqual(current, r.Before) {
		return empty, errCLIInstallationConflict
	}
	proposed := before
	proposed.DeviceCredentials = append([]AIAgentClientDeviceCredentialSnapshot(nil), before.DeviceCredentials...)
	if r.State == "complete" {
		for i := range proposed.DeviceCredentials {
			if proposed.DeviceCredentials[i].DeviceID == r.Before.DeviceID {
				proposed.DeviceCredentials[i] = r.After
			}
		}
		if err := trial.restoreSnapshot(proposed); err != nil {
			return empty, err
		}
	}
	proposed.ExpectedVersion, proposed.Version, proposed.SavedAt = s.version, s.version+1, s.snapshotNow().UTC()
	ctx = withDeviceEnrollmentMutation(ctx, before, r.Before.DeviceID)
	ctx = context.WithValue(ctx, deviceBootstrapMutationKey{}, mutation)
	ctx = context.WithValue(ctx, taskAgentPreserveMembershipContextKey{}, true)
	writeErr := s.snapshotStore.SaveAIAgentClientSnapshot(ctx, proposed)
	// Even an error after commit may be a lost response. Re-read BOTH the
	// transition and current native authority; the credential alone is not proof
	// that admission and the durable acknowledgement committed atomically.
	readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	committed, found, err := s.snapshotStore.(deviceBootstrapRepository).readDeviceBootstrap(readCtx, r.Request.TransitionID)
	if err == nil && found && reflect.DeepEqual(committed, r) {
		_, err = s.currentDeviceBootstrap(readCtx, committed, nil)
	} else if err == nil {
		err = errCLIInstallationConflict
	}
	if err != nil {
		s.snapshotReloadMu.Lock()
		s.snapshotReloadRequired, s.snapshotLastReloadAt = true, time.Time{}
		s.snapshotReloadMu.Unlock()
		if writeErr != nil {
			return empty, writeErr
		}
		return empty, err
	}
	s.publishDeviceEnrollment(trial, r.Before.DeviceID)
	s.version = proposed.Version
	s.markSnapshotObserved(proposed.SavedAt)
	return committed, nil
}
