package riidoaiserver

import "time"

// Unknown stays absent. Historical rows do not record source completeness, so
// different hashes alone cannot distinguish an edit from a partial projection.
func runReadModifiedSinceRead(read AssignmentContextRead, current runReadMetadata) *bool {
	if read.Status != "served" && read.Status != "truncated" {
		return nil
	}
	before, beforeErr := time.Parse(time.RFC3339Nano, normalizedTaskContextTime(read.SourceUpdatedAt))
	after, afterErr := time.Parse(time.RFC3339Nano, normalizedTaskContextTime(current.source.UpdatedAt))
	timesComparable := beforeErr == nil && afterErr == nil
	if timesComparable && after.After(before) {
		return runReadModified(true)
	}
	if read.Status != "truncated" && current.comparable && read.SourceVersion != "" && current.source.Version != "" {
		if read.SourceVersion == current.source.Version {
			return runReadModified(false)
		}
	}
	return nil
}

func runReadModified(value bool) *bool { return &value }
