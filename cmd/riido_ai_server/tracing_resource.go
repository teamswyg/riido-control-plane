package main

import (
	"runtime/debug"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Container builds set this from the workflow source SHA without exporting env.
var buildRevision = ""

func tracingResource(serviceName string) *resource.Resource {
	return resource.NewSchemaless(
		attribute.String("service.name", serviceName),
		attribute.String("service.version", tracingSourceRevision(buildRevision, debug.ReadBuildInfo)),
	)
}

func tracingSourceRevision(revision string, readInfo func() (*debug.BuildInfo, bool)) string {
	if validSourceRevision(revision) {
		return revision
	}
	if revision != "" {
		return "unknown"
	}
	info, ok := readInfo()
	if !ok || info == nil {
		return "unknown"
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.modified" && setting.Value == "true" {
			return "unknown"
		}
		if setting.Key == "vcs.revision" {
			revision = setting.Value
		}
	}
	if validSourceRevision(revision) {
		return revision
	}
	return "unknown"
}

func validSourceRevision(revision string) bool {
	if len(revision) != 40 {
		return false
	}
	for _, c := range revision {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
