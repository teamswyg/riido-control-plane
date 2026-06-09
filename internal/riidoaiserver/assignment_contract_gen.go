package riidoaiserver

import assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

const SchemaVersion = assignmentcontract.SchemaVersion

type AssignmentState = assignmentcontract.AssignmentState

const (
	AssignmentQueued     = assignmentcontract.AssignmentQueued
	AssignmentLeased     = assignmentcontract.AssignmentLeased
	AssignmentReady      = assignmentcontract.AssignmentReady
	AssignmentRunning    = assignmentcontract.AssignmentRunning
	AssignmentCancelling = assignmentcontract.AssignmentCancelling
	AssignmentCancelled  = assignmentcontract.AssignmentCancelled
	AssignmentCompleted  = assignmentcontract.AssignmentCompleted
	AssignmentFailed     = assignmentcontract.AssignmentFailed
)

type PollAction = assignmentcontract.PollAction

const (
	PollNone   = assignmentcontract.PollNone
	PollStart  = assignmentcontract.PollStart
	PollCancel = assignmentcontract.PollCancel
	PollActive = assignmentcontract.PollActive
)

const (
	EventAssignmentQueued       = assignmentcontract.EventAssignmentQueued
	EventAssignmentLeased       = assignmentcontract.EventAssignmentLeased
	EventAssignmentReady        = assignmentcontract.EventAssignmentReady
	EventAssignmentRunning      = assignmentcontract.EventAssignmentRunning
	EventAssignmentCancelling   = assignmentcontract.EventAssignmentCancelling
	EventAssignmentCancelled    = assignmentcontract.EventAssignmentCancelled
	EventAssignmentCompleted    = assignmentcontract.EventAssignmentCompleted
	EventAssignmentFailed       = assignmentcontract.EventAssignmentFailed
	EventAssignmentStateUpdated = assignmentcontract.EventAssignmentStateUpdated
	EventRiidoLog               = assignmentcontract.EventRiidoLog
	EventProviderSessionPinned  = assignmentcontract.EventProviderSessionPinned
	EventProviderLog            = assignmentcontract.EventProviderLog
	EventProviderWarning        = assignmentcontract.EventProviderWarning
	EventProviderError          = assignmentcontract.EventProviderError
)

func isAgentActive(state AssignmentState) bool {
	return assignmentcontract.IsAgentActive(state)
}

func isTerminal(state AssignmentState) bool {
	return assignmentcontract.IsTerminal(state)
}

func canTransitionAssignment(from, to AssignmentState) bool {
	return assignmentcontract.CanTransition(from, to)
}
