package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	var texts []string
	for _, path := range paths {
		record, text, err := scanWorkflow(root, path, accepted, used)
		if err != nil {
			return auditResult{}, err
		}
		texts = append(texts, text)
		result.Records = append(result.Records, record)
		addRecord(&result, record)
	}
	result.EvidenceTools, result.EvidenceToolCovered, result.MissingEvidenceTools = auditEvidenceTools(root, texts)
	for path := range accepted {
		if !used[path] {
			result.AcceptedUnused = append(result.AcceptedUnused, path)
		}
	}
	sort.Strings(result.AcceptedUnused)
	return result, nil
}

func scanWorkflow(root, path string, accepted map[string]acceptedGap, used map[string]bool) (workflowRecord, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return workflowRecord{}, "", fmt.Errorf("read workflow: %w", err)
	}
	rel, text := slashPath(path[len(root)+1:]), string(data)
	uploadModes := artifactUploadModes(text)
	evidenceOut := evidenceOutPaths(text)
	uploadPaths := artifactUploadPathValues(text)
	record := workflowRecord{
		Path:                 rel,
		HasExecutable:        hasExecutableStep(text),
		HasEvidenceOut:       len(evidenceOut) > 0,
		EvidenceOutCount:     len(evidenceOut),
		UploadedEvidenceOut:  countUploadedEvidenceOut(evidenceOut, uploadPaths),
		MissingEvidenceOut:   missingEvidenceUploads(evidenceOut, uploadPaths),
		UploadsArtifact:      len(uploadModes) > 0,
		ArtifactUploadCount:  len(uploadModes),
		StrictUploadCount:    countUploadMode(uploadModes, "error"),
		NonStrictUploadCount: countNonStrictUploadModes(uploadModes),
	}
	return classify(record, accepted, used), text, nil
}

func acceptedByPath(gaps []acceptedGap) map[string]acceptedGap {
	out := make(map[string]acceptedGap, len(gaps))
	for _, gap := range gaps {
		out[slashPath(gap.Path)] = gap
	}
	return out
}
