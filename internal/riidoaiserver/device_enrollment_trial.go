package riidoaiserver

import (
	"context"
	"time"
)

func (s *PersistentAIAgentClientStore) enrollmentTrial() (*DevelopmentAIAgentClientStore, AIAgentClientSnapshot, error) {
	before, err := s.snapshot(s.snapshotNow().UTC())
	if err != nil {
		return nil, before, err
	}
	trial := NewDevelopmentAIAgentClientStore()
	if err := trial.restoreSnapshot(before); err != nil {
		return nil, before, err
	}
	s.mu.Lock()
	trial.expectedDaemonProfile, trial.daemonClientPolicy = s.expectedDaemonProfile, s.daemonClientPolicy
	trial.executionCoverageEnrollment = s.executionCoverageEnrollment
	_, trial.executionCoverageAtomic = s.snapshotStore.(executionCoverageEnrollmentStore)
	s.mu.Unlock()
	return trial, before, nil
}

func (s *PersistentAIAgentClientStore) commitDeviceEnrollment(ctx context.Context, trial *DevelopmentAIAgentClientStore, before AIAgentClientSnapshot, response EnrollDeviceResponse) (EnrollDeviceResponse, error) {
	snapshot, err := trial.snapshot(s.snapshotNow().UTC())
	if err != nil {
		return EnrollDeviceResponse{}, err
	}
	snapshot.ExpectedVersion, snapshot.Version = s.version, s.version+1
	ctx = withDeviceEnrollmentMutation(ctx, before, response.DeviceID)
	ctx = context.WithValue(ctx, taskAgentPreserveMembershipContextKey{}, true)
	if err := s.snapshotStore.SaveAIAgentClientSnapshot(ctx, snapshot); err != nil {
		return s.resolveDeviceEnrollment(ctx, trial, snapshot, response, err)
	}
	s.publishDeviceEnrollment(trial, response.DeviceID)
	s.version = snapshot.Version
	s.markSnapshotObserved(snapshot.SavedAt)
	return response, nil
}

func (s *PersistentAIAgentClientStore) resolveDeviceEnrollment(ctx context.Context, trial *DevelopmentAIAgentClientStore, proposed AIAgentClientSnapshot, response EnrollDeviceResponse, writeErr error) (EnrollDeviceResponse, error) {
	s.snapshotReloadMu.Lock()
	s.snapshotLastReloadAt = time.Time{}
	s.snapshotReloadRequired = true
	s.snapshotReloadMu.Unlock()
	readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	committed, found, err := s.snapshotStore.LoadAIAgentClientSnapshot(readCtx)
	if err != nil || !found || !sameEnrollmentCredential(committed, proposed, response.DeviceID) {
		return EnrollDeviceResponse{}, writeErr
	}
	if response.ExecutionCoverageGeneration != "" {
		authority, ok := s.snapshotStore.(executionCoverageEnrollmentStore)
		if !ok || authority.verifyEnrollmentCoverage(readCtx, response) != nil {
			return EnrollDeviceResponse{}, writeErr
		}
	}
	s.publishDeviceEnrollment(trial, response.DeviceID)
	s.version = committed.Version
	s.snapshotReloadMu.Lock()
	s.snapshotReloadRequired = false
	s.snapshotReloadMu.Unlock()
	s.markSnapshotObserved(committed.SavedAt)
	return response, nil
}
