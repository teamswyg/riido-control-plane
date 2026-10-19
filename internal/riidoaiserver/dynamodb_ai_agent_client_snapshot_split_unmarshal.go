package riidoaiserver

import (
	"encoding/json"
	"fmt"
)

func unmarshalDynamoDBAIAgentClientSnapshotPart(name string, raw []byte, snapshot *AIAgentClientSnapshot) error {
	switch name {
	case dynamoDBAIAgentClientSnapshotPartOnboardingAgentCreates:
		return json.Unmarshal(raw, &snapshot.OnboardingAgentCreates)
	case dynamoDBAIAgentClientSnapshotPartTaskRequestReceipts:
		return json.Unmarshal(raw, &snapshot.TaskRequestReceipts)
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
	case dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships:
		return json.Unmarshal(raw, &snapshot.TaskAgentMemberships)
	case dynamoDBAIAgentClientSnapshotPartTaskThreads:
		return json.Unmarshal(raw, &snapshot.TaskThreads)
	case dynamoDBAIAgentClientSnapshotPartTaskThreadMessages:
		return json.Unmarshal(raw, &snapshot.TaskThreadMessages)
	case dynamoDBAIAgentClientSnapshotPartEvents:
		return json.Unmarshal(raw, &snapshot.Events)
	case dynamoDBAIAgentClientSnapshotPartOnboardingCreated:
		return json.Unmarshal(raw, &snapshot.OnboardingCreatedIDs)
	case dynamoDBAIAgentClientSnapshotPartOnboardingDone:
		return json.Unmarshal(raw, &snapshot.OnboardingCompletions)
	case dynamoDBAIAgentClientSnapshotPartOnboardingThreads:
		return json.Unmarshal(raw, &snapshot.OnboardingCompletionThreads)
	default:
		return fmt.Errorf("unknown AI Agent client snapshot part %s", name)
	}
}
