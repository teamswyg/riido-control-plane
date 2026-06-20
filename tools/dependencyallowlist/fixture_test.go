package main

func testContract() contract {
	return contract{
		SchemaVersion: schemaVersion,
		Service:       "riido-control-plane",
		Policy:        "test",
		AllowedDirectModules: []allowedModule{
			{Path: "github.com/teamswyg/riido-contracts", Layer: "contract", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel/sdk", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
			{Path: "go.opentelemetry.io/otel/trace", Layer: "observability", Owner: "platform", Approved: true, Reason: "test"},
		},
	}
}

func testModules() []goModule {
	return []goModule{
		{Path: "github.com/teamswyg/riido-control-plane", Main: true},
		{Path: "github.com/teamswyg/riido-contracts", Version: "v0.3.6"},
		{Path: "go.opentelemetry.io/otel", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/sdk", Version: "v1.44.0"},
		{Path: "go.opentelemetry.io/otel/trace", Version: "v1.44.0"},
		{Path: "google.golang.org/grpc", Version: "v1.81.1", Indirect: true},
	}
}
