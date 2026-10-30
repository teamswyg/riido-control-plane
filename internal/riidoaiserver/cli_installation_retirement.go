package riidoaiserver

import (
	"errors"
	"time"
)

const cliInstallationRetirementSchema = "riido-cli-installation-link-intent-retirement.v1"

type cliInstallationRetirementRequest cliInstallationReceiptRequest

type cliInstallationRetirement struct {
	cliInstallationRetirementRequest
	Status  string                  `json:"status"`
	Receipt *cliInstallationReceipt `json:"receipt,omitempty"`
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
	result := cliInstallationRetirement{cliInstallationRetirementRequest: req, Status: "retired"}
	if intent.Consumed {
		receipt, err := originalCLIInstallationReceipt(intent, link)
		if err != nil {
			return cliInstallationRetirement{}, err
		}
		result.Status, result.Receipt = "consumed", &receipt
	}
	return result, nil
}
