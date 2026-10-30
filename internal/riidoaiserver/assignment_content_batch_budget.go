package riidoaiserver

import "context"

// A batched content read returns consecutive captured pages of one part so an
// agent reads a large value in a few calls instead of one call per 8 KiB page.
// It never crosses into another part and always includes the requested page.
const (
	assignmentContentBatchMinBytes = 8 << 10
	assignmentContentBatchMaxBytes = 256 << 10
)

func validAssignmentContentBatchBytes(view string, maxBytes int) bool {
	if maxBytes == 0 {
		return true
	}
	return view == AssignmentContextViewContent && maxBytes >= assignmentContentBatchMinBytes && maxBytes <= assignmentContentBatchMaxBytes
}

// AssignmentContentRangeLoader is an optional store capability: it reads each
// physical block once while visiting consecutive logical pages. visit returns
// false to stop. Stores without it are read one page at a time.
type AssignmentContentRangeLoader interface {
	VisitAssignmentContentPages(ctx context.Context, assignmentID string, first, count int, visit func(AssignmentContentPage) bool) error
}

func uniqueAssignmentContentSources(pages []AssignmentContentPage) []AssignmentContextSource {
	seen := map[AssignmentContextSource]bool{}
	out := []AssignmentContextSource{}
	for _, page := range pages {
		for _, source := range page.Sources {
			if !seen[source] {
				seen[source] = true
				out = append(out, source)
			}
		}
	}
	return out
}
