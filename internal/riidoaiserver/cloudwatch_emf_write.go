package riidoaiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func WriteCloudWatchEMF(w io.Writer, config CloudWatchEMFConfig, snapshot MetricsSnapshot) error {
	if w == nil {
		return errors.New("riidoaiserver: metrics writer is required")
	}
	envelope := buildCloudWatchEMFEnvelope(normalizeCloudWatchEMFConfig(config), snapshot)
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(body))
	return err
}
