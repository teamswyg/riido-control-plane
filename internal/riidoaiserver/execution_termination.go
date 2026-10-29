package riidoaiserver

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"
)

const (
	executionTerminationSchema = "riido-execution-termination.v1"
	executionTerminationKind   = "local_invocations_quiescent"
	executionLeaseHeader       = "X-Riido-Assignment-Lease-Token"
)

var (
	errTerminationForbidden   = errors.New("execution termination proof is not authorized")
	errTerminationConflict    = errors.New("execution termination observation conflicts")
	errTerminationUnavailable = errors.New("execution termination persistence is unavailable")
)

type executionTerminationRequest struct {
	SchemaVersion      string `json:"schema_version"`
	Kind               string `json:"kind"`
	StartedInvocations uint64 `json:"started_invocations"`
}

type executionTerminationReceipt struct {
	SchemaVersion      string    `json:"schema_version"`
	Kind               string    `json:"kind"`
	AssignmentID       string    `json:"assignment_id"`
	LeaseEpoch         string    `json:"lease_epoch"`
	StartedInvocations uint64    `json:"started_invocations"`
	RecordedAt         time.Time `json:"recorded_at"`
}

type executionTerminationIdentity struct {
	deviceID, ownerPrincipalID, assignmentID, epoch string
	coverage                                        executionCoverageAuth
}

type executionTerminationRecorder interface {
	recordExecutionTermination(context.Context, executionTerminationIdentity, executionTerminationRequest) (executionTerminationReceipt, error)
}

func (id executionTerminationIdentity) key() string { return id.assignmentID + "#" + id.epoch }

func (id executionTerminationIdentity) matches(claim executionClaim) bool {
	return claim.DeviceID == id.deviceID && claim.OwnerPrincipalID == id.ownerPrincipalID && claim.AssignmentID == id.assignmentID && subtle.ConstantTimeCompare([]byte(claim.LeaseEpoch), []byte(id.epoch)) == 1
}

func newExecutionTermination(id executionTerminationIdentity, req executionTerminationRequest, at time.Time) executionTerminationReceipt {
	return executionTerminationReceipt{req.SchemaVersion, req.Kind, id.assignmentID, id.epoch, req.StartedInvocations, at.UTC()}
}

func matchingExecutionTermination(receipt executionTerminationReceipt, req executionTerminationRequest) (executionTerminationReceipt, error) {
	if receipt.SchemaVersion != req.SchemaVersion || receipt.Kind != req.Kind || receipt.StartedInvocations != req.StartedInvocations {
		return executionTerminationReceipt{}, errTerminationConflict
	}
	return receipt, nil
}
