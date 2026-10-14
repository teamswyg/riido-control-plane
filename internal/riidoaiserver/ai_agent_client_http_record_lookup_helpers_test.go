package riidoaiserver

func findAIAgent(agents []AgentClientRecord, id string) (AgentClientRecord, bool) {
	for _, agent := range agents {
		if agent.AgentID == id {
			return agent, true
		}
	}
	return AgentClientRecord{}, false
}

func findDevice(devices []DeviceRecord, id string) (DeviceRecord, bool) {
	for _, device := range devices {
		if device.DeviceID == id {
			return device, true
		}
	}
	return DeviceRecord{}, false
}

func findRuntime(runtimes []RuntimeRecord, id string) (RuntimeRecord, bool) {
	for _, runtime := range runtimes {
		if runtime.RuntimeID == id {
			return runtime, true
		}
	}
	return RuntimeRecord{}, false
}

func runtimeIsMarkedAssigned(devices []DeviceRecord, runtimeID string) bool {
	for _, device := range devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID == runtimeID {
				return runtime.HasAssignedAgent
			}
		}
	}
	return false
}
