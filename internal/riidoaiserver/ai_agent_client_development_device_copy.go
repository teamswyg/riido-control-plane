package riidoaiserver

import (
	"slices"
	"strings"
)

func copyDevices(devices []DeviceRecord) []DeviceRecord {
	out := make([]DeviceRecord, 0, len(devices))
	for _, device := range devices {
		out = append(out, copyDevice(device))
	}
	return out
}

func copyDevice(device DeviceRecord) DeviceRecord {
	runtimes := make([]RuntimeRecord, len(device.Runtimes))
	for i, runtime := range device.Runtimes {
		runtimes[i] = copyRuntime(runtime)
	}
	device.Runtimes = runtimes
	device.ConnectedWorkspaceIDs = append([]string(nil), device.ConnectedWorkspaceIDs...)
	if device.ClientStatus != nil {
		status := *device.ClientStatus
		device.ClientStatus = &status
	}
	return device
}

func addConnectedWorkspace(connected []string, workspaceID string) []string {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return connected
	}
	if slices.Contains(connected, workspaceID) {
		return connected
	}
	return append(connected, workspaceID)
}

func deviceConnectedToWorkspace(device DeviceRecord, workspaceID string) bool {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return false
	}
	return slices.Contains(device.ConnectedWorkspaceIDs, workspaceID)
}

func copyRuntime(runtime RuntimeRecord) RuntimeRecord {
	runtime.Models = append([]RuntimeModelRecord(nil), runtime.Models...)
	return runtime
}

func copyDeviceDaemon(daemon DeviceDaemonRecord) DeviceDaemonRecord {
	daemon.SupportedActions = append([]DaemonControlAction(nil), daemon.SupportedActions...)
	return daemon
}
