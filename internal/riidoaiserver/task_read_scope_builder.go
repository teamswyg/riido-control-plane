package riidoaiserver

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

type taskReadScopeBuilder struct {
	capture assignmentReadScopeCapture
	nonce   string
	budget  taskReadScopePackingBudget
}

func newTaskReadScopeBuilder() (*taskReadScopeBuilder, error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return &taskReadScopeBuilder{nonce: base64.RawURLEncoding.EncodeToString(nonce)}, nil
}

func (b *taskReadScopeBuilder) addPart(resource AssignmentScopedResource, part AssignmentContentPart, version string, pages []AssignmentContentPage, complete bool) AssignmentContentManifest {
	manifest := AssignmentContentManifest{ResourceKind: resource.Kind, SourceID: resource.ID, Part: part, Version: version, Status: "partial", Cursors: []string{}}
	for i, page := range pages {
		page.ResourceKind, page.SourceID, page.Part, page.Version = resource.Kind, resource.ID, part, version
		canonical, err := json.Marshal([]any{"riido-assignment-content/v1", b.nonce, resource.Kind, resource.ID, part, version, page.Offset})
		if err != nil {
			complete = false
			break
		}
		digest := sha256.Sum256(canonical)
		page.Cursor = base64.RawURLEncoding.EncodeToString(digest[:])
		page.Complete = complete && i == len(pages)-1
		if validateAssignmentContentPage(page) != nil {
			complete = false
			break
		}
		if !b.budget.appendPage(page) {
			complete = false
			break
		}
		b.capture.Pages = append(b.capture.Pages, page)
		manifest.Cursors = append(manifest.Cursors, page.Cursor)
	}
	manifest.Complete = complete && len(pages) > 0 && len(manifest.Cursors) == len(pages)
	if manifest.Complete {
		manifest.Status = "available"
	}
	if len(manifest.Cursors) > 0 {
		manifest.NextAction = "riido context part " + resource.ID + " --cursor " + manifest.Cursors[0]
	}
	b.capture.Manifest = append(b.capture.Manifest, manifest)
	return manifest
}
