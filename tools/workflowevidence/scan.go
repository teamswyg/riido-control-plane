package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func auditWorkflows(root string, m manifest) (auditResult, error) {
	paths, err := filepath.Glob(repoPath(root, filepath.Join(m.WorkflowRoot, "*.yml")))
	if err != nil {
		return auditResult{}, err
	}
	sort.Strings(paths)
	accepted := acceptedByPath(m.AcceptedGaps)
	used := map[string]bool{}
	var result auditResult
	for _, path := range paths {
		record, err := scanWorkflow(root, path, accepted, used)
		if err != nil {
			return auditResult{}, err
		}
		result.Records = append(result.Records, record)
		addRecord(&result, record)
	}
	for path := range accepted {
		if !used[path] {
			result.AcceptedUnused = append(result.AcceptedUnused, path)
		}
	}
	sort.Strings(result.AcceptedUnused)
	return result, nil
}

func scanWorkflow(root, path string, accepted map[string]acceptedGap, used map[string]bool) (workflowRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return workflowRecord{}, fmt.Errorf("read workflow: %w", err)
	}
	rel, text := slashPath(path[len(root)+1:]), string(data)
	uploadModes := artifactUploadModes(text)
	record := workflowRecord{
		Path:                 rel,
		HasExecutable:        hasExecutableStep(text),
		HasEvidenceOut:       strings.Contains(text, "evidence-out"),
		UploadsArtifact:      len(uploadModes) > 0,
		ArtifactUploadCount:  len(uploadModes),
		StrictUploadCount:    countUploadMode(uploadModes, "error"),
		NonStrictUploadCount: countNonStrictUploadModes(uploadModes),
	}
	return classify(record, accepted, used), nil
}

func acceptedByPath(gaps []acceptedGap) map[string]acceptedGap {
	out := make(map[string]acceptedGap, len(gaps))
	for _, gap := range gaps {
		out[slashPath(gap.Path)] = gap
	}
	return out
}
