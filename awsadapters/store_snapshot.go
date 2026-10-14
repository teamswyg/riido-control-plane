package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type (
	StoreSnapshot               = internal.StoreSnapshot
	StoreSnapshotTask           = internal.StoreSnapshotTask
	DynamoDBStoreSnapshotConfig = internal.DynamoDBStoreSnapshotConfig
	DynamoDBStoreSnapshot       = internal.DynamoDBStoreSnapshot
)

var NewDynamoDBStoreSnapshot = internal.NewDynamoDBStoreSnapshot
