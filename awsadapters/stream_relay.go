package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type (
	StreamRelayPublisher       = internal.StreamRelayPublisher
	StreamRelayCheckpointStore = internal.StreamRelayCheckpointStore
	StreamRelayEvent           = internal.StreamRelayEvent
	StreamRelayCheckpoint      = internal.StreamRelayCheckpoint
	DynamoDBStreamRelayConfig  = internal.DynamoDBStreamRelayConfig
	DynamoDBStreamRelay        = internal.DynamoDBStreamRelay
	DynamoDBStreamRelayStats   = internal.DynamoDBStreamRelayStats

	DynamoDBStreamRelayCheckpointStoreConfig = internal.DynamoDBStreamRelayCheckpointStoreConfig
	DynamoDBStreamRelayCheckpointStore       = internal.DynamoDBStreamRelayCheckpointStore
)

var (
	NewDynamoDBStreamRelay                = internal.NewDynamoDBStreamRelay
	RunDynamoDBStreamRelayOnce            = internal.RunDynamoDBStreamRelayOnce
	NewDynamoDBStreamRelayCheckpointStore = internal.NewDynamoDBStreamRelayCheckpointStore
)
