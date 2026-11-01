package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type capturedPartDigestPage struct {
	Offset   int             `json:"offset"`
	Status   string          `json:"status"`
	Complete bool            `json:"complete"`
	Data     json.RawMessage `json:"data"`
}

func assignmentCapturedPartDigest(resource AssignmentScopedResource, part AssignmentContentManifest, pages []AssignmentContentPage) (string, error) {
	// Identity, title and selection reasons are meaningful metadata. Delivery
	// cursors and capture time are not source content, and are deliberately absent.
	metadata := resource
	metadata.Parts = nil
	canonicalPages := make([]capturedPartDigestPage, 0, len(pages))
	for _, page := range pages {
		data, err := canonicalCapturedJSON(page.Data)
		if err != nil {
			return "", err
		}
		canonicalPages = append(canonicalPages, capturedPartDigestPage{page.Offset, page.Status, page.Complete, data})
	}
	input := struct {
		Schema   string                   `json:"schema"`
		Resource AssignmentScopedResource `json:"resource"`
		Part     AssignmentContentPart    `json:"part"`
		Version  string                   `json:"version"`
		Status   string                   `json:"status"`
		Complete bool                     `json:"complete"`
		Pages    []capturedPartDigestPage `json:"pages"`
	}{"riido-captured-part-digest.v1", metadata, part.Part, part.Version, part.Status, part.Complete, canonicalPages}
	raw, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
