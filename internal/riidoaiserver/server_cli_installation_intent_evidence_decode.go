package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func decodeCLIInstallationIntentEvidence(w http.ResponseWriter, r *http.Request) (cliInstallationIntentCreateRequest, error) {
	var wire cliInstallationIntentCreateRequest
	bounded := http.MaxBytesReader(w, r.Body, 4<<10)
	data, err := io.ReadAll(bounded)
	_ = bounded.Close()
	if err != nil {
		var oversized *http.MaxBytesError
		if !errors.As(err, &oversized) {
			err = errCLIInstallationInvalid
		}
		return wire, err
	}
	var discriminator struct {
		SchemaVersion string `json:"schema_version"`
	}
	if json.Unmarshal(data, &discriminator) != nil {
		return wire, errCLIInstallationInvalid
	}
	keys := "schema_version request_id installation_id resource client_id key_fingerprint public_key_algorithm public_key expected_device_id approved_principal_id"
	switch discriminator.SchemaVersion {
	case cliInstallationIntentEvidenceSchema:
	case cliInstallationTransferEvidenceSchema:
		keys += " previous_device_id previous_link_revision"
	default:
		return wire, errCLIInstallationInvalid
	}
	// The schema probe selects a closed key set; only the strict decoder accepts it.
	r.Body = io.NopCloser(bytes.NewReader(data))
	err = decodeCLIInstallation(w, r, keys, &wire)
	return wire, err
}
