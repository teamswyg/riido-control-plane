package riidoaiserver

func runtimeSelectionFromDevices(devices []DeviceRecord, runtimeID string, requestedModelID *string) (RuntimeKind, RuntimeModelRecord, bool) {
	for _, device := range devices {
		for _, runtime := range device.Runtimes {
			if runtime.RuntimeID != runtimeID {
				continue
			}
			model, ok := runtimeModelSelection(runtime, requestedModelID)
			return runtime.Kind, model, ok
		}
	}
	return "", RuntimeModelRecord{}, false
}
