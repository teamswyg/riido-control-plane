package riidoaiserver

func eventDeviceRecord(payload any) (DeviceRecord, bool) {
	switch event := payload.(type) {
	case DeviceRuntimeSnapshotEvent:
		return event.Device, true
	default:
		return DeviceRecord{}, false
	}
}
