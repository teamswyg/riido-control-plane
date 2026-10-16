package main

import (
	"context"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracingSourceRevisionRejectsUnboundedOrDirtyValues(t *testing.T) {
	const revision = "82787a78f678300ce7dc9faafde4bca190ba56b9"
	for _, tc := range []struct {
		name     string
		built    string
		settings []debug.BuildSetting
		want     string
	}{
		{"container revision", revision, nil, revision},
		{"clean local revision", "", []debug.BuildSetting{{Key: "vcs.revision", Value: revision}, {Key: "vcs.modified", Value: "false"}}, revision},
		{"dirty local revision", "", []debug.BuildSetting{{Key: "vcs.revision", Value: revision}, {Key: "vcs.modified", Value: "true"}}, "unknown"},
		{"invalid build override", "private-token", []debug.BuildSetting{{Key: "vcs.revision", Value: revision}}, "unknown"},
		{"invalid embedded revision", "", []debug.BuildSetting{{Key: "vcs.revision", Value: "private-url"}}, "unknown"},
		{"unversioned", "", nil, "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tracingSourceRevision(tc.built, func() (*debug.BuildInfo, bool) { return &debug.BuildInfo{Settings: tc.settings}, true })
			if got != tc.want {
				t.Fatalf("source revision = %q, want %q", got, tc.want)
			}
		})
	}
	if got := tracingSourceRevision("", func() (*debug.BuildInfo, bool) { return nil, false }); got != "unknown" {
		t.Fatal("unavailable build metadata claimed a source")
	}
}

func TestTraceExporterCarriesBuiltSourceInServiceVersion(t *testing.T) {
	const revision = "82787a78f678300ce7dc9faafde4bca190ba56b9"
	previous := buildRevision
	buildRevision = revision
	t.Cleanup(func() { buildRevision = previous })
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter), sdktrace.WithResource(tracingResource("riido-test")))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	_, span := (otelTraceRecorder{tracer: provider.Tracer("test")}).StartTraceSpan(context.Background(), riidoaiserver.TraceSpanStart{Name: "test"})
	span.End()
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatal("completed span was not exported")
	}
	attributes := spans.Snapshots()[0].Resource().Attributes()
	found := false
	for _, attr := range attributes {
		if string(attr.Key) == "service.version" {
			found = attr.Value.AsString() == revision
		}
		if strings.Contains(attr.Value.AsString(), "private") {
			t.Fatal("unexpected resource metadata")
		}
	}
	if !found {
		t.Fatal("exported span lacks the built source revision")
	}
}
