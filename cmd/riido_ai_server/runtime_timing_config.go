package main

import "time"

type runtimeTimingConfig struct {
	ShutdownTimeout       time.Duration
	MetricsLogInterval    time.Duration
	AssignmentActiveLease time.Duration
	LongPollMaxHold       time.Duration
	LongPollTick          time.Duration
}

func runtimeTimingFromEnv() (runtimeTimingConfig, error) {
	shutdownTimeout, err := envDurationSeconds(envShutdownTimeoutSeconds, 10*time.Second)
	if err != nil {
		return runtimeTimingConfig{}, err
	}
	metricsLogInterval, err := envOptionalDurationSeconds(envMetricsLogInterval)
	if err != nil {
		return runtimeTimingConfig{}, err
	}
	assignmentActiveLease, err := envOptionalDurationSeconds(envAssignmentActiveLease)
	if err != nil {
		return runtimeTimingConfig{}, err
	}
	longPollMaxHold, err := envDurationSeconds(envLongPollMaxHoldSeconds, 25*time.Second)
	if err != nil {
		return runtimeTimingConfig{}, err
	}
	longPollTick, err := envDurationSeconds(envLongPollTickSeconds, 2*time.Second)
	if err != nil {
		return runtimeTimingConfig{}, err
	}
	return runtimeTimingConfig{
		ShutdownTimeout:       shutdownTimeout,
		MetricsLogInterval:    metricsLogInterval,
		AssignmentActiveLease: assignmentActiveLease,
		LongPollMaxHold:       longPollMaxHold,
		LongPollTick:          longPollTick,
	}, nil
}
