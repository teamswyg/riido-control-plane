package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type (
	TaskEvent            = internal.TaskEvent
	OutboxRecord         = internal.OutboxRecord
	DynamoDBOutboxConfig = internal.DynamoDBOutboxConfig
	DynamoDBOutbox       = internal.DynamoDBOutbox
)

var NewDynamoDBOutbox = internal.NewDynamoDBOutbox
