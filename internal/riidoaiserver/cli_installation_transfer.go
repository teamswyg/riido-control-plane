package riidoaiserver

import (
	"math"
	"strconv"
)

const cliInstallationIntentV3Schema = "riido-cli-installation-link-intent.v3"

func (req cliInstallationIntentRequest) isTransfer() bool {
	return req.SchemaVersion == cliInstallationIntentV3Schema
}

func cliInstallationLinkRevision(raw string) (uint64, error) {
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 || strconv.FormatUint(n, 10) != raw {
		return 0, errCLIInstallationInvalid
	}
	return n, nil
}

func validCLIInstallationLinkRevision(raw string) bool {
	_, err := cliInstallationLinkRevision(raw)
	return err == nil
}

func validCLIInstallationTransfer(req cliInstallationIntentRequest) bool {
	if !req.isTransfer() {
		return req.SchemaVersion == cliInstallationIntentSchema && req.PreviousDeviceID == "" && req.PreviousLinkRevision == ""
	}
	n, err := cliInstallationLinkRevision(req.PreviousLinkRevision)
	return err == nil && n < math.MaxUint64 && cliInstallationID(req.PreviousDeviceID) && req.PreviousDeviceID != req.ExpectedDeviceID
}

func expectedCLIInstallationLinkRevision(req cliInstallationIntentRequest) string {
	if !req.isTransfer() {
		return "1"
	}
	n, err := cliInstallationLinkRevision(req.PreviousLinkRevision)
	if err != nil || n == math.MaxUint64 {
		return ""
	}
	return strconv.FormatUint(n+1, 10)
}
