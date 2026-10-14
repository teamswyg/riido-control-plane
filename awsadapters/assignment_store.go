package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type (
	AssignmentState                        = internal.AssignmentState
	Assignment                             = internal.Assignment
	AssignmentOperationRecord              = internal.AssignmentOperationRecord
	AssignmentClaimResult                  = internal.AssignmentClaimResult
	AssignmentActiveLease                  = internal.AssignmentActiveLease
	AssignmentProjection                   = internal.AssignmentProjection
	DynamoDBAssignmentOperationStoreConfig = internal.DynamoDBAssignmentOperationStoreConfig
	DynamoDBAssignmentOperationStore       = internal.DynamoDBAssignmentOperationStore
)

var NewDynamoDBAssignmentOperationStore = internal.NewDynamoDBAssignmentOperationStore
