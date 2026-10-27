package riidoaiserver

import (
	"context"
	"errors"
	"strconv"
	"time"
)

func (s *Store) readCLIInstallationPolicy(ctx context.Context, id cliInstallationIdentity, owners cliInstallationDeviceOwnerReader) (cliInstallationPolicy, error) {
	if s.snapshotStore != nil && s.operationStore != nil {
		return cliInstallationPolicy{}, errCLIInstallationUnavailable
	}
	if durable, ok := s.operationStore.(cliInstallationRepository); ok {
		return durable.readCLIInstallationPolicy(ctx, id, owners)
	}
	for range 3 {
		result, err := s.executeCLIInstallation(ctx, cliInstallationCommand{read: &cliInstallationRead{identity: id}})
		if err != nil {
			return cliInstallationPolicy{}, err
		}
		owner := ""
		if result.read.found {
			owner, err = owners.cliInstallationDeviceOwner(ctx, result.read.link.DeviceID)
			if err != nil {
				return cliInstallationPolicy{}, err
			}
		}
		_, err = s.executeCLIInstallation(ctx, cliInstallationCommand{read: &cliInstallationRead{identity: id, expected: &result.read}})
		if err == nil {
			return installationPolicyFromRead(result.read, owner, s.now()), nil
		}
		if !errors.Is(err, errCLIInstallationConflict) {
			return cliInstallationPolicy{}, err
		}
	}
	return cliInstallationPolicy{}, errCLIInstallationUnavailable
}

func (s *Store) readLocalCLIInstallation(state *storeState, req cliInstallationRead) cliInstallationResult {
	link, found := state.cliInstallationLinks[req.identity.key()]
	if found && link.cliInstallationIdentity != req.identity {
		return cliInstallationResult{err: ErrAuthorizationForbidden}
	}
	set := cliInstallationReadSet{link: link, found: found, revision: state.devicePolicyRevisions[link.DeviceID], observation: deviceExecutionUnknown}
	if req.expected != nil {
		if set.link != req.expected.link || set.found != req.expected.found || set.revision != req.expected.revision {
			return cliInstallationResult{err: errCLIInstallationConflict}
		}
		return cliInstallationResult{read: set}
	}
	if found && set.revision == 0 {
		return cliInstallationResult{err: errCLIInstallationUnavailable}
	}
	for _, claim := range state.executionClaims {
		if claim.DeviceID == link.DeviceID && claimMatchesNonOwnerAssignment(claim, state.assignments[claim.AssignmentID]) {
			set.observation = deviceExecutionBlocked
			break
		}
	}
	return cliInstallationResult{read: set}
}

func installationPolicyFromRead(set cliInstallationReadSet, owner string, now time.Time) cliInstallationPolicy {
	result := cliInstallationPolicy{SchemaVersion: cliInstallationPolicySchema, Decision: "unknown", LinkRevision: "0", ExecutionRevision: "0", CoverageState: "unknown", CheckedAt: now.UTC()}
	if set.found {
		result.LinkRevision, result.ExecutionRevision = set.link.LinkRevision, strconv.FormatUint(set.revision, 10)
		if owner != "" && owner == set.link.DeviceOwnerAtLink && owner == set.link.LinkApprovedBy && set.observation == deviceExecutionBlocked {
			result.Decision = "blocked"
		}
	}
	return result
}
