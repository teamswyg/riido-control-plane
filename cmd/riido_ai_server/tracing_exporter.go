package main

import (
	"net/url"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func otelTraceHTTPExporterOptions(endpoint string) []otlptracehttp.Option {
	parsed, err := url.Parse(endpoint)
	if err == nil && parsed.Scheme == "" {
		return []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
		}
	}
	options := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(endpoint)}
	if err == nil && parsed.Scheme == "http" {
		options = append(options, otlptracehttp.WithInsecure())
	}
	return options
}
