package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

type metricsReaderFunc func(context.Context) (MetricsSnapshot, error)

func (f metricsReaderFunc) Metrics(ctx context.Context) (MetricsSnapshot, error) {
	return f(ctx)
}

type metricsContributorFunc func(MetricsSnapshot) MetricsSnapshot

func (f metricsContributorFunc) ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot {
	return f(snapshot)
}

type pointerMetricsReader struct{}

func (*pointerMetricsReader) Metrics(context.Context) (MetricsSnapshot, error) {
	return MetricsSnapshot{}, nil
}

func TestObservedMetricsReaderFiltersNilAndAppliesContributors(t *testing.T) {
	base := metricsReaderFunc(func(context.Context) (MetricsSnapshot, error) {
		return MetricsSnapshot{SchemaVersion: MetricsSchemaVersion, TasksTotal: 1}, nil
	})
	reader := NewObservedMetricsReader(base, nil, metricsContributorFunc(func(snapshot MetricsSnapshot) MetricsSnapshot {
		snapshot.TasksTotal++
		snapshot.AssignmentsTotal = 3
		return snapshot
	}))
	got, err := reader.Metrics(context.Background())
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if got.TasksTotal != 2 || got.AssignmentsTotal != 3 {
		t.Fatalf("snapshot = %+v", got)
	}
}

func TestObservedMetricsReaderKeepsNilAndErrorBehavior(t *testing.T) {
	if NewObservedMetricsReader(nil) != nil {
		t.Fatal("nil base reader should remain nil")
	}
	base := &pointerMetricsReader{}
	if NewObservedMetricsReader(base, nil) != base {
		t.Fatal("base reader should be reused when contributors are empty")
	}
	want := errors.New("metrics failed")
	reader := NewObservedMetricsReader(metricsReaderFunc(func(context.Context) (MetricsSnapshot, error) {
		return MetricsSnapshot{}, want
	}), metricsContributorFunc(func(snapshot MetricsSnapshot) MetricsSnapshot {
		t.Fatal("contributor should not run after base error")
		return snapshot
	}))
	if _, err := reader.Metrics(context.Background()); !errors.Is(err, want) {
		t.Fatalf("Metrics err = %v, want %v", err, want)
	}
}
