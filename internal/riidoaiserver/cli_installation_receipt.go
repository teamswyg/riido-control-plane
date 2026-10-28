package riidoaiserver

import "time"

const cliInstallationReceiptSchema = "riido-cli-installation-link-receipt.v1"

type cliInstallationReceiptRequest struct {
	SchemaVersion string `json:"schema_version"`
	IntentID      string `json:"intent_id"`
	cliInstallationIdentity
	ExpectedDeviceID    string `json:"expected_device_id"`
	ApprovedPrincipalID string `json:"approved_principal_id"`
}

type cliInstallationReceipt struct {
	SchemaVersion string              `json:"schema_version"`
	IntentID      string              `json:"intent_id"`
	Link          cliInstallationLink `json:"link"`
}

func (req cliInstallationReceiptRequest) valid() bool {
	return req.SchemaVersion == cliInstallationReceiptSchema && req.cliInstallationIdentity.valid() &&
		cliInstallationID(req.IntentID) && cliInstallationID(req.ExpectedDeviceID) && cliInstallationID(req.ApprovedPrincipalID)
}

func checkCLIInstallationReceiptIntent(req cliInstallationReceiptRequest, intent cliInstallationIntent, now time.Time) error {
	if !req.valid() {
		return errCLIInstallationInvalid
	}
	if intent.valid() != nil {
		return errCLIInstallationUnavailable
	}
	if intent.ID != req.IntentID || intent.Request.cliInstallationIdentity != req.cliInstallationIdentity ||
		intent.Request.ExpectedDeviceID != req.ExpectedDeviceID || intent.Request.ApprovedPrincipalID != req.ApprovedPrincipalID {
		return ErrAuthorizationForbidden
	}
	if !intent.Consumed {
		if !now.Before(intent.ExpiresAt) {
			return errCLIInstallationExpired
		}
		return errCLIInstallationConflict
	}
	return nil
}

func originalCLIInstallationReceipt(intent cliInstallationIntent, link cliInstallationLink) (cliInstallationReceipt, error) {
	// A receipt proves the historical link only. Current owner and execution
	// permission are independently checked by the policy endpoint.
	if link.valid() != nil || link.cliInstallationIdentity != intent.Request.cliInstallationIdentity ||
		link.DeviceID != intent.Request.ExpectedDeviceID || link.LinkApprovedBy != intent.Request.ApprovedPrincipalID ||
		link.DeviceOwnerAtLink != intent.Request.ApprovedPrincipalID || !link.LinkedAt.Before(intent.ExpiresAt) {
		return cliInstallationReceipt{}, errCLIInstallationUnavailable
	}
	return cliInstallationReceipt{cliInstallationReceiptSchema, intent.ID, link}, nil
}
