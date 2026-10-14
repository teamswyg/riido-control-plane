package riidoaiserver

func eventDeviceDaemon(payload any) (DeviceDaemonRecord, bool) {
	switch event := payload.(type) {
	case DeviceDaemonStatusEvent:
		return event.Daemon, true
	default:
		return DeviceDaemonRecord{}, false
	}
}
