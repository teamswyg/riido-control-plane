package riidoaiserver

import assignmentcontract "github.com/teamswyg/riido-contracts/assignment"

const SchemaVersion = assignmentcontract.SchemaVersion

type AssignmentState = assignmentcontract.AssignmentState
type AssignmentStateCode = assignmentcontract.AssignmentStateCode

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

const (
	AssignmentStateCodeQueued     = assignmentcontract.AssignmentStateCodeQueued
	AssignmentStateCodeLeased     = assignmentcontract.AssignmentStateCodeLeased
	AssignmentStateCodeReady      = assignmentcontract.AssignmentStateCodeReady
	AssignmentStateCodeRunning    = assignmentcontract.AssignmentStateCodeRunning
	AssignmentStateCodeCancelling = assignmentcontract.AssignmentStateCodeCancelling
	AssignmentStateCodeCancelled  = assignmentcontract.AssignmentStateCodeCancelled
	AssignmentStateCodeCompleted  = assignmentcontract.AssignmentStateCodeCompleted
	AssignmentStateCodeFailed     = assignmentcontract.AssignmentStateCodeFailed
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
	EventProviderLog            = assignmentcontract.EventProviderLog
	EventProviderWarning        = assignmentcontract.EventProviderWarning
	EventProviderError          = assignmentcontract.EventProviderError
)

func isAgentActive(state AssignmentState) bool {
	return state.Code().IsAgentActive()
}

func isTerminal(state AssignmentState) bool {
	return state.Code().IsTerminal()
}

func canTransitionAssignment(from, to AssignmentState) bool {
	return assignmentcontract.CanTransitionCode(from.Code(), to.Code())
}
