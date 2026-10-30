package riidoaiserver

import (
	"errors"
	"time"
)

const cliInstallationRetirementSchema = "riido-cli-installation-link-intent-retirement.v1"

type cliInstallationRetirementRequest cliInstallationReceiptRequest

type cliInstallationRetirement struct {
	SchemaVersion string `json:"schema_version"`
	IntentID      string `json:"intent_id"`
	Status        string `json:"status"`
	// A consumed receipt already carries this complete binding. Duplicating
	// it can exceed the HTTP budget for otherwise valid opaque identities.
	InstallationID      string                  `json:"installation_id,omitempty"`
	Resource            string                  `json:"resource,omitempty"`
	ClientID            string                  `json:"client_id,omitempty"`
	KeyFingerprint      string                  `json:"key_fingerprint,omitempty"`
	ExpectedDeviceID    string                  `json:"expected_device_id,omitempty"`
	ApprovedPrincipalID string                  `json:"approved_principal_id,omitempty"`
	Receipt             *cliInstallationReceipt `json:"receipt,omitempty"`
}

func (req cliInstallationRetirementRequest) valid() bool {
	receipt := cliInstallationReceiptRequest(req)
	receipt.SchemaVersion = cliInstallationReceiptSchema
	return req.SchemaVersion == cliInstallationRetirementSchema && receipt.valid()
}

func checkCLIInstallationRetirement(req cliInstallationRetirementRequest, intent cliInstallationIntent, now time.Time) error {
	if !req.valid() {
		return errCLIInstallationInvalid
	}
	receipt := cliInstallationReceiptRequest(req)
	receipt.SchemaVersion = cliInstallationReceiptSchema
	err := checkCLIInstallationReceiptIntent(receipt, intent, now)
	if errors.Is(err, errCLIInstallationExpired) {
		return nil // Only an expired unconsumed intent can be retired.
	}
	return err // Consumed returns its original receipt; live pending is conflict.
}

func cliInstallationRetirementResult(req cliInstallationRetirementRequest, intent cliInstallationIntent, link cliInstallationLink) (cliInstallationRetirement, error) {
	result := cliInstallationRetirement{SchemaVersion: req.SchemaVersion, IntentID: req.IntentID, Status: "retired"}
	if intent.Consumed {
		receipt, err := originalCLIInstallationReceipt(intent, link)
		if err != nil {
			return cliInstallationRetirement{}, err
		}
		result.Status, result.Receipt = "consumed", &receipt
	} else {
		result.InstallationID, result.Resource = req.InstallationID, req.Resource
		result.ClientID, result.KeyFingerprint = req.ClientID, req.KeyFingerprint
		result.ExpectedDeviceID, result.ApprovedPrincipalID = req.ExpectedDeviceID, req.ApprovedPrincipalID
	}
	return result, nil
}
