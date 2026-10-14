package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTracingSampleRatio = 0.01
	defaultTracingServiceName = "riido_ai_server"
	tracingShutdownTimeout    = 5 * time.Second
)

type tracingRuntimeConfig struct {
	Enabled      bool
	SampleRatio  float64
	OTLPEndpoint string
	ServiceName  string
}

func tracingConfigFromEnv() (tracingRuntimeConfig, error) {
	enabled, err := envOptionalBool(envTracingEnabled)
	if err != nil {
		return tracingRuntimeConfig{}, err
	}
	sampleRatio, err := envOptionalFloat64(envTracingSampleRatio, defaultTracingSampleRatio)
	if err != nil {
		return tracingRuntimeConfig{}, err
	}
	if sampleRatio < 0 || sampleRatio > 1 {
		return tracingRuntimeConfig{}, fmt.Errorf("%s must be between 0 and 1", envTracingSampleRatio)
	}
	return tracingRuntimeConfig{
		Enabled:      enabled,
		SampleRatio:  sampleRatio,
		OTLPEndpoint: tracingEndpointFromEnv(),
		ServiceName:  tracingServiceNameFromEnv(),
	}, nil
}

func envOptionalFloat64(key string, fallback float64) (float64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", key)
	}
	return value, nil
}
