package riidoaiserver

import "context"

const (
	cliInstallationIntentEvidenceSchema   = "riido-cli-installation-intent-evidence.v1"
	cliInstallationTransferEvidenceSchema = "riido-cli-installation-intent-evidence.v2"
)

type cliInstallationIntentEvidence struct {
	SchemaVersion string                               `json:"schema_version"`
	RequestID     string                               `json:"request_id"`
	Found         bool                                 `json:"found"`
	Intent        *cliInstallationIntentReplayResponse `json:"intent,omitempty"`
}

type cliInstallationIntentEvidenceReader interface {
	readCLIInstallationIntentEvidence(context.Context, cliInstallationIntentCreateRequest) (cliInstallationIntentEvidence, error)
}

// This lookup has no candidate nonce or expiry and must never enter create/replay.
func installationEvidenceLookup(wire cliInstallationIntentCreateRequest) (cliInstallationIntentReplay, error) {
	var schema string
	switch wire.SchemaVersion {
	case cliInstallationIntentEvidenceSchema:
		schema = cliInstallationIntentV2Schema
	case cliInstallationTransferEvidenceSchema:
		schema = cliInstallationIntentV3Schema
	default:
		return cliInstallationIntentReplay{}, errCLIInstallationInvalid
	}
	wire.SchemaVersion = schema
	req, id, err := normalizeCLIInstallationIntentCreate(wire)
	req.SchemaVersion = schema
	return cliInstallationIntentReplay{Request: req, Response: cliInstallationIntentReplayResponse{RequestID: id}}, err
}

func installationEvidenceResponse(lookup cliInstallationIntentReplay, original cliInstallationIntentReplayResponse, found bool) cliInstallationIntentEvidence {
	result := cliInstallationIntentEvidence{SchemaVersion: cliInstallationIntentEvidenceSchema, RequestID: lookup.Response.RequestID, Found: found}
	if lookup.Request.isTransfer() {
		result.SchemaVersion = cliInstallationTransferEvidenceSchema
	}
	if found {
		result.Intent = &original
	}
	return result
}

func readInstallationEvidence(snapshot cliInstallationSnapshot, lookup cliInstallationIntentReplay) (cliInstallationIntentEvidence, error) {
	original, found := snapshot.IntentReplays[lookup.key()]
	if !found {
		if _, orphan := snapshot.Intents[lookup.intentID()]; orphan {
			return cliInstallationIntentEvidence{}, errCLIInstallationUnavailable
		}
		return installationEvidenceResponse(lookup, cliInstallationIntentReplayResponse{}, false), nil
	}
	intent, exists := snapshot.Intents[original.Response.IntentID]
	if original.valid() != nil || original.key() != lookup.key() || !exists || intent.valid() != nil || !original.matches(intent) {
		return cliInstallationIntentEvidence{}, errCLIInstallationUnavailable
	}
	if original.Request != lookup.Request || original.Response.RequestID != lookup.Response.RequestID {
		return cliInstallationIntentEvidence{}, errCLIInstallationConflict
	}
	return installationEvidenceResponse(lookup, original.Response, true), nil
}
