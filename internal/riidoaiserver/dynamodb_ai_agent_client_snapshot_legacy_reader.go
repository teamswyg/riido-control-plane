package riidoaiserver

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

func legacyDynamoDBAIAgentClientSnapshotReader(item map[string]map[string]string) (io.Reader, int64, error) {
	if gzipped := dynamoDBStringValue(item, "snapshot_gzip"); gzipped != "" {
		raw, err := gunzipBase64(gzipped)
		if err != nil {
			return nil, 0, fmt.Errorf("decode DynamoDB AI Agent client snapshot gzip: %w", err)
		}
		return bytes.NewReader(raw), int64(len(gzipped)), nil
	}
	rawSnapshot := dynamoDBStringValue(item, "snapshot_json")
	if rawSnapshot == "" {
		return nil, 0, errors.New("decode DynamoDB AI Agent client snapshot response: snapshot_gzip or snapshot_json is required")
	}
	return strings.NewReader(rawSnapshot), int64(len(rawSnapshot)), nil
}
