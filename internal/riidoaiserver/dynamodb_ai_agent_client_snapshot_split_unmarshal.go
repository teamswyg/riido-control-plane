package riidoaiserver

import (
	"encoding/json"
	"fmt"
)

func unmarshalDynamoDBAIAgentClientSnapshotPart(name string, raw []byte, snapshot *AIAgentClientSnapshot) error {
	switch name {
	case dynamoDBAIAgentClientSnapshotPartDevices:
		return json.Unmarshal(raw, &snapshot.Devices)
	case dynamoDBAIAgentClientSnapshotPartDeviceCredentials:
		return json.Unmarshal(raw, &snapshot.DeviceCredentials)
	case dynamoDBAIAgentClientSnapshotPartDeviceConnections:
		return json.Unmarshal(raw, &snapshot.DeviceConnectionGrants)
	case dynamoDBAIAgentClientSnapshotPartDaemons:
		return json.Unmarshal(raw, &snapshot.Daemons)
	case dynamoDBAIAgentClientSnapshotPartAgents:
		return json.Unmarshal(raw, &snapshot.Agents)
	case dynamoDBAIAgentClientSnapshotPartFixtures:
		return json.Unmarshal(raw, &snapshot.Fixtures)
	case dynamoDBAIAgentClientSnapshotPartTaskThreads:
		return json.Unmarshal(raw, &snapshot.TaskThreads)
	case dynamoDBAIAgentClientSnapshotPartTaskThreadMessages:
		return json.Unmarshal(raw, &snapshot.TaskThreadMessages)
	case dynamoDBAIAgentClientSnapshotPartEvents:
		return json.Unmarshal(raw, &snapshot.Events)
	default:
		return fmt.Errorf("unknown AI Agent client snapshot part %s", name)
	}
}
