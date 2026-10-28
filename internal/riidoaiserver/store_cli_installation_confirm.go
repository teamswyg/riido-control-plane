package riidoaiserver

func (s *Store) confirmLocalCLIInstallation(state *storeState, req cliInstallationConfirmation) (cliInstallationLink, error) {
	intent, ok := state.cliInstallationIntents[req.id]
	if !ok || intent.Request.Resource != req.resource {
		return cliInstallationLink{}, ErrAuthorizationForbidden
	}
	if err := verifyCLIInstallationProof(intent, req.device, req.signature); err != nil {
		return cliInstallationLink{}, err
	}
	link, exists := state.cliInstallationLinks[intent.Request.key()]
	if exists && !matchingCLIInstallationLink(link, intent, req.device) {
		return cliInstallationLink{}, errCLIInstallationConflict
	}
	if intent.Consumed {
		if !exists {
			return cliInstallationLink{}, errCLIInstallationUnavailable
		}
		return link, nil
	}
	if !s.now().Before(intent.ExpiresAt) {
		return cliInstallationLink{}, errCLIInstallationExpired
	}
	if !exists {
		link = linkForCLIInstallation(intent, req.device, s.now())
		if err := link.valid(); err != nil {
			return cliInstallationLink{}, err
		}
		if err := bumpDevicePolicyRevision(state, req.device.deviceID); err != nil {
			return cliInstallationLink{}, err
		}
		state.cliInstallationLinks[intent.Request.key()] = link
	}
	intent.Consumed = true
	state.cliInstallationIntents[req.id] = intent
	return link, nil
}
