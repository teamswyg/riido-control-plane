package riidoaiserver

import "sort"

func copyDeviceConnectionGrants(in map[string]map[string]DeviceConnectionGrant) []DeviceConnectionGrant {
	out := make([]DeviceConnectionGrant, 0)
	for _, deviceGrants := range in {
		for _, grant := range deviceGrants {
			out = append(out, grant)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return deviceConnectionGrantKey(out[i].DeviceID, out[i].PrincipalID, out[i].WorkspaceID) <
			deviceConnectionGrantKey(out[j].DeviceID, out[j].PrincipalID, out[j].WorkspaceID)
	})
	return out
}
