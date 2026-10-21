package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"unicode/utf8"
)

const maxTaskDocumentSnapshotBytes = 256 * 1024

// AIAgentTaskDocumentSnapshot is an execution input, never a document-history write.
type AIAgentTaskDocumentSnapshot struct {
	Content       string `json:"content"`
	ContentFormat string `json:"content_format"`
}

func (s *AIAgentTaskDocumentSnapshot) UnmarshalJSON(data []byte) error {
	var raw struct {
		Content       *string `json:"content"`
		ContentFormat *string `json:"content_format"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	if raw.Content == nil || raw.ContentFormat == nil {
		return errors.New("document_snapshot requires content and content_format")
	}
	s.Content, s.ContentFormat = *raw.Content, *raw.ContentFormat
	return s.validate()
}

func (s *AIAgentTaskDocumentSnapshot) validate() error {
	if s == nil {
		return nil
	}
	if s.ContentFormat != "html" {
		return errors.New("document_snapshot content_format must be html")
	}
	if len(s.Content) > maxTaskDocumentSnapshotBytes || !utf8.ValidString(s.Content) {
		return errors.New("document_snapshot content must be valid UTF-8 of at most 262144 bytes")
	}
	return nil
}

func (s *AIAgentTaskDocumentSnapshot) apply(document *AIAgentTaskContextDocument) {
	if s == nil {
		return
	}
	document.Content, document.ContentFormat, document.Source = s.Content, s.ContentFormat, "editor_snapshot"
}
