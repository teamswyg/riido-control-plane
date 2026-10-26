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
}

// assignmentContextSources lists every source carried by a task context, in
// the vocabulary of the read-record contract (task, project, milestone,
// parent_task, link, attachment; part body/title/metadata).
func assignmentContextSources(c AIAgentTaskContext) []AssignmentContextSource {
	sources := []AssignmentContextSource{{Kind: "task", ID: c.Component.ID, Part: "body", Version: assignmentContextVersion(c.Document.Content)}}
	sources = append(sources, assignmentContextResourceSources(c.Resources)...)
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
	source := AssignmentContextSource{Kind: kind, ID: ref.ID, Part: "title"}
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
