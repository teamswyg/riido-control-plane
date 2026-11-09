package riidoaiserver

import (
	"context"
	"math"
)

func (s *PersistentAIAgentClientStore) advanceDeviceBootstrap(ctx context.Context, id, phase string, device deviceBootstrapAuth, proof deviceBootstrapProof) (deviceBootstrapRecord, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	record, err := s.loadDeviceBootstrap(ctx, id, device)
	if err != nil {
		return deviceBootstrapRecord{}, err
	}
	// Recovery returns the same committed record only under the current native
	// credential. It never replays an old authority over a later rotation.
	if (phase == "prepare" && record.State == "prepared") || (phase == "activate" && record.State == "complete") {
		return record, nil
	}
	if record.Revision == math.MaxUint64 {
		return deviceBootstrapRecord{}, errCLIInstallationConflict
	}
	if err := verifyDeviceBootstrapProof(record, phase, proof, s.snapshotNow().UTC()); err != nil {
		return deviceBootstrapRecord{}, err
	}
	next := record
	next.Revision++
	if phase == "prepare" {
		revision, err := s.snapshotStore.(deviceBootstrapRepository).readBootstrapLegacyRevision(ctx, record.Before.DeviceID)
		if err != nil {
			return deviceBootstrapRecord{}, err
		}
		if revision >= math.MaxUint64-1 {
			return deviceBootstrapRecord{}, errCLIInstallationConflict
		}
		next.State, next.PreviousBootDigest, next.PolicyRevision = "prepared", proof.BootDigest, revision+1
	} else {
		if record.PolicyRevision == math.MaxUint64 {
			return deviceBootstrapRecord{}, errCLIInstallationConflict
		}
		next.State, next.CompletedBootDigest, next.CompletedAt = "complete", proof.BootDigest, s.snapshotNow().UTC()
		next.PolicyRevision++
	}
	return s.commitDeviceBootstrap(ctx, deviceBootstrapMutation{before: &record, after: next})
}
