package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type AssignmentContextSource struct {
	Kind    string `json:"kind"`
	ID      string `json:"id,omitempty"`
	Part    string `json:"part"`
	Version string `json:"version,omitempty"`
	// UpdatedAt is the source's modification time when the server read it.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// assignmentContextSources lists every source in a task context using the
// read-record vocabulary (task, project, milestone, parent_task, link, attachment).
func assignmentContextSources(c AIAgentTaskContext) []AssignmentContextSource {
	sources := append([]AssignmentContextSource{assignmentContextTaskSource(c, "body")}, assignmentContextResourceSources(c.Resources)...)
	for _, ref := range assignmentContextHierarchyRefs(c.Hierarchy) {
		sources = append(sources, assignmentContextReferenceSources(ref.kind, ref.ref)...)
	}
	return sources
}

type assignmentContextHierarchyRef struct {
	kind string
	ref  AIAgentTaskContextReference
}

func assignmentContextHierarchyRefs(h AIAgentTaskContextHierarchy) []assignmentContextHierarchyRef {
	var refs []assignmentContextHierarchyRef
	for _, ref := range []assignmentContextHierarchyRef{{"project", h.Project}, {"milestone", h.Milestone}, {"parent_task", h.ParentTask}} {
		if strings.TrimSpace(ref.ref.ID) != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

func assignmentContextReferenceSources(kind string, ref AIAgentTaskContextReference) []AssignmentContextSource {
	source := AssignmentContextSource{Kind: kind, ID: ref.ID, Part: "title", UpdatedAt: normalizedTaskContextTime(ref.UpdatedAt)}
	if ref.Document != nil {
		source.Part, source.Version = "body", assignmentContextVersion(ref.Document.Content)
	}
	return append([]AssignmentContextSource{source}, assignmentContextResourceSources(ref.Resources)...)
}

func assignmentContextResourceSources(resources []AIAgentTaskContextResource) []AssignmentContextSource {
	sources := make([]AssignmentContextSource, 0, len(resources))
	for _, resource := range resources {
		kind := "link"
		if resource.Kind == "file" {
			kind = "attachment"
		}
		id := resource.ID
		if id == "" {
			id = assignmentContextVersion(resource.URL)
		}
		sources = append(sources, AssignmentContextSource{Kind: kind, ID: id, Part: "metadata"})
	}
	return sources
}

// assignmentContextVersion is a content hash, so a read record names the
// exact text the agent saw even when the document later changes.
func assignmentContextVersion(content string) string {
	if content == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(sum[:8])
}
