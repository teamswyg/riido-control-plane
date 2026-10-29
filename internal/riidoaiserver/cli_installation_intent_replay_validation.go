package riidoaiserver

import "encoding/json"

func normalizeCLIInstallationIntentCreate(wire cliInstallationIntentCreateRequest) (cliInstallationIntentRequest, string, error) {
	req := wire.cliInstallationIntentRequest
	var requestID string
	switch req.SchemaVersion {
	case cliInstallationIntentSchema:
		if len(wire.RequestID) != 0 {
			return req, "", errCLIInstallationInvalid
		}
	case cliInstallationIntentV2Schema:
		if json.Unmarshal(wire.RequestID, &requestID) != nil || !installationRandom32(requestID) {
			return req, "", errCLIInstallationInvalid
		}
		req.SchemaVersion = cliInstallationIntentSchema
	default:
		return req, "", errCLIInstallationInvalid
	}
	_, err := installationPublicKey(req)
	return req, requestID, err
}

func (record cliInstallationIntentReplay) valid() error {
	req := record.Request
	if req.SchemaVersion != cliInstallationIntentV2Schema || record.Response.SchemaVersion != cliInstallationIntentV2Schema {
		return errCLIInstallationInvalid
	}
	req.SchemaVersion = cliInstallationIntentSchema
	if _, err := installationPublicKey(req); err != nil {
		return err
	}
	if !installationRandom32(record.Response.RequestID) || record.Response.IntentID != record.intentID() || !installationRandom32(record.Response.Challenge) || record.Response.ExpiresAt.IsZero() {
		return errCLIInstallationInvalid
	}
	data, err := json.Marshal(record)
	if err != nil || len(data) > 3<<10 {
		return errCLIInstallationInvalid
	}
	return nil
}
