package riidoaiserver

func checkCLIInstallationTransferLink(intent cliInstallationIntent, link cliInstallationLink, device authenticatedClaimDevice) error {
	if intent.Consumed {
		if matchingCLIInstallationLink(link, intent, device) {
			return nil
		}
		return errCLIInstallationConflict
	}
	r := intent.Request
	if link.cliInstallationIdentity != r.cliInstallationIdentity || link.DeviceID != r.PreviousDeviceID || link.LinkRevision != r.PreviousLinkRevision || link.DeviceOwnerAtLink != r.ApprovedPrincipalID || link.LinkApprovedBy != r.ApprovedPrincipalID {
		return errCLIInstallationConflict
	}
	return nil
}
