package riidoaiserver

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func gzipBase64(b []byte) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(b); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func gunzipBase64(s string) ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return io.ReadAll(gz)
}

func decodeAIAgentClientSnapshot(r io.Reader) (AIAgentClientSnapshot, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var snapshot AIAgentClientSnapshot
	if err := dec.Decode(&snapshot); err != nil {
		return AIAgentClientSnapshot{}, fmt.Errorf("decode AI Agent client snapshot: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return AIAgentClientSnapshot{}, errors.New("decode AI Agent client snapshot: trailing data")
	}
	return snapshot, nil
}
