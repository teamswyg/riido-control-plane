package main

import "testing"

func TestExecutionCoverageEnrollmentDefaultOffAndRejectsUnsupportedAuthority(t *testing.T) {
	t.Setenv(envExecutionCoverageEnrollment, "")
	if enabled, err := executionCoverageEnrollmentFromEnv(nil, nil); enabled || err != nil {
		t.Fatal("coverage enrollment is not default-off")
	}
	t.Setenv(envExecutionCoverageEnrollment, "true")
	if enabled, err := executionCoverageEnrollmentFromEnv(nil, nil); enabled || err == nil {
		t.Fatal("coverage enrollment accepted non-atomic authority")
	}
	t.Setenv(envExecutionCoverageEnrollment, "invalid")
	if _, err := executionCoverageEnrollmentFromEnv(nil, nil); err == nil {
		t.Fatal("invalid flag accepted")
	}
}
