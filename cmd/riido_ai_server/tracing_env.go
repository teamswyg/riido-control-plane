package main

import (
	"os"
	"strings"
)

func tracingEndpointFromEnv() string {
	endpoint := strings.TrimSpace(os.Getenv(envTracingOTLPEndpoint))
	if endpoint != "" {
		return endpoint
	}
	return strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
}

func tracingServiceNameFromEnv() string {
	serviceName := strings.TrimSpace(os.Getenv(envTracingServiceName))
	if serviceName == "" {
		serviceName = strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	}
	if serviceName == "" {
		return defaultTracingServiceName
	}
	return serviceName
}
