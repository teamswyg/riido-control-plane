package riidoaiserver

import (
	"errors"
	"strings"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *Store) validateRecoveryEvent(state *storeState, current Assignment, req AgentEventRequest) (AgentEventRequest, error) {
	context := assignment.RecoveryContextFromMetadata(req.Metadata)
	if context == nil {
		return req, nil
	}
	key := strings.TrimSpace(req.Metadata[metadatakeys.AssignmentEventKey.String()])
	if req.EventType != EventRiidoLog || key == "" || len(key) > 256 || req.DaemonID == "" || req.DeviceID == "" || req.RuntimeID == "" {
		return req, errors.New("recovery event requires scoped execution identity and event key")
	}
	if current.LeaseToken == "" || req.Metadata["lease_token"] != current.LeaseToken {
		return req, errors.New("recovery event execution lease does not match")
	}
	if req.State != "" && req.State != current.State {
		return req, errors.New("recovery event must not change assignment state")
	}
	if req.ProviderSessionID != "" && req.ProviderSessionID != current.ProviderSessionID {
		return req, errors.New("recovery event must not change provider session")
	}
	if isTerminal(current.State) || current.State == AssignmentCancelling {
		if context.State != "cancelled" {
			return req, errors.New("recovery event belongs to a stopped execution")
		}
	} else {
		if current.State == AssignmentQueued {
			return req, errors.New("recovery event requires an active execution")
		}
		expired, err := s.assignmentActiveLeaseExpired(state, current, s.now())
		if err != nil {
			return req, err
		}
		if expired {
			return req, errors.New("recovery event execution lease expired")
		}
	}
	if context.Provider != current.RuntimeProvider && context.Provider != "unknown" {
		allowed := false
		if current.RecoveryPolicy != nil && current.ResumeSessionID == "" && current.ProviderSessionID == "" {
			for _, candidate := range current.RecoveryPolicy.Fallbacks {
				if candidate.Provider == context.Provider {
					allowed = true
				}
			}
		}
		if !allowed {
			return req, errors.New("recovery event provider is outside accepted policy")
		}
		if context.State == "checking" || context.State == "switching" {
			if err := s.recheckRecoveryCandidate(current, context.Provider); err != nil {
				return req, err
			}
		}
	}
	req.Metadata = assignment.SanitizeRecoveryMetadata(req.Metadata)
	metadata := context.Metadata()
	metadata[metadatakeys.AssignmentEventKey.String()] = key
	req.Metadata = metadata // The recovery journal accepts only bounded fields and its dedup key.
	req.Message = ""        // Recovery diagnostics never publish provider output as a client progress line.
	return req, nil
}
