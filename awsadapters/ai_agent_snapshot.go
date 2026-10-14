package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type (
	AIAgentClientSnapshot                 = internal.AIAgentClientSnapshot
	AIAgentClientDeviceCredentialSnapshot = internal.AIAgentClientDeviceCredentialSnapshot
	AIAgentClientEventSnapshot            = internal.AIAgentClientEventSnapshot
	DynamoDBAIAgentClientSnapshotConfig   = internal.DynamoDBAIAgentClientSnapshotConfig
	DynamoDBAIAgentClientSnapshot         = internal.DynamoDBAIAgentClientSnapshot
)

var NewDynamoDBAIAgentClientSnapshot = internal.NewDynamoDBAIAgentClientSnapshot
