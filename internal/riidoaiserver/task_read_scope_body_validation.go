package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"unicode/utf8"
)

func validateReadScopeBody(page *taskReadBodyPage, resourceID, part, version string, offset int) error {
	if page == nil || page.ResourceID != resourceID || page.Part != part || page.Offset == nil || *page.Offset != offset || page.Complete == nil || page.Source == "" || (page.Format != "html" && page.Format != "json" && page.Format != "text") || !utf8.ValidString(page.Content) || page.Version == "" || (version != "" && page.Version != version) {
		return taskContextReadError{"invalid"}
	}
	if *page.Complete {
		if page.NextOffset != nil || (page.Status != "available" && page.Status != "empty") {
			return taskContextReadError{"invalid"}
		}
	} else if page.NextOffset == nil || *page.NextOffset != offset+len(page.Content) || *page.NextOffset <= offset || (page.Status != "partial" && page.Status != "available") {
		return taskContextReadError{"invalid"}
	}
	return nil
}

func readScopeBodyVersion(kind AssignmentContentKind, id, part, source, format, content string) string {
	canonical, _ := json.Marshal([]string{string(kind), id, part, source, format, content}) //nolint:errchkjson // A string slice cannot fail JSON encoding.
	hash := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func readScopeBodyPart(part string) AssignmentContentPart {
	switch part {
	case "document":
		return AssignmentContentEditor
	case "description":
		return AssignmentContentDescription
	case "document_alternate":
		return AssignmentContentAlternate
	}
	return ""
}
