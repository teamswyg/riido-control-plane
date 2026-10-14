package riidoaiserver

import "context"

type observedMetricsReader struct {
	base         MetricsReader
	contributors []MetricsSnapshotContributor
}

type MetricsSnapshotContributor interface {
	ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot
}

func NewObservedMetricsReader(base MetricsReader, contributors ...MetricsSnapshotContributor) MetricsReader {
	if base == nil {
		return base
	}
	filtered := make([]MetricsSnapshotContributor, 0, len(contributors))
	for _, contributor := range contributors {
		if contributor != nil {
			filtered = append(filtered, contributor)
		}
	}
	if len(filtered) == 0 {
		return base
	}
	return observedMetricsReader{base: base, contributors: filtered}
}

func (r observedMetricsReader) Metrics(ctx context.Context) (MetricsSnapshot, error) {
	snapshot, err := r.base.Metrics(ctx)
	if err != nil {
		return MetricsSnapshot{}, err
	}
	for _, contributor := range r.contributors {
		snapshot = contributor.ApplyToMetricsSnapshot(snapshot)
	}
	return snapshot, nil
}
