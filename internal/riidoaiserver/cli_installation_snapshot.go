package riidoaiserver

import "maps"

type cliInstallationSnapshot struct {
	Intents         map[string]cliInstallationIntent `json:"intents,omitempty"`
	Links           map[string]cliInstallationLink   `json:"links,omitempty"`
	DeviceRevisions map[string]uint64                `json:"device_revisions,omitempty"`
}

func cliInstallationSnapshotFromState(state *storeState) cliInstallationSnapshot {
	return cliInstallationSnapshot{maps.Clone(state.cliInstallationIntents), maps.Clone(state.cliInstallationLinks), maps.Clone(state.devicePolicyRevisions)}
}

func restoreCLIInstallations(state *storeState, snapshot cliInstallationSnapshot) error {
	for id, intent := range snapshot.Intents {
		if id != intent.ID {
			return errCLIInstallationInvalid
		}
		if err := intent.valid(); err != nil {
			return err
		}
		state.cliInstallationIntents[id] = intent
	}
	for id, link := range snapshot.Links {
		if id != link.key() {
			return errCLIInstallationInvalid
		}
		if err := link.valid(); err != nil {
			return err
		}
		if snapshot.DeviceRevisions[link.DeviceID] == 0 {
			return errCLIInstallationUnavailable
		}
		state.cliInstallationLinks[id] = link
	}
	for id, revision := range snapshot.DeviceRevisions {
		if !cliInstallationID(id) || revision == 0 || revision < state.devicePolicyRevisions[id] {
			return errCLIInstallationUnavailable
		}
		state.devicePolicyRevisions[id] = revision
	}
	for _, intent := range state.cliInstallationIntents {
		if intent.Consumed {
			link, ok := state.cliInstallationLinks[intent.Request.key()]
			device := authenticatedClaimDevice{deviceID: intent.Request.ExpectedDeviceID, ownerPrincipalID: intent.Request.ApprovedPrincipalID}
			if !ok || !matchingCLIInstallationLink(link, intent, device) {
				return errCLIInstallationUnavailable
			}
		}
	}
	return nil
}
