package main

import "testing"

func clearRiidoAIServerEnv(t *testing.T) {
	t.Helper()
	for _, key := range riidoAIServerEnvKeys() {
		t.Setenv(key, "")
	}
}

func riidoAIServerEnvKeys() []string {
	return []string{
		envAddr, envShutdownTimeoutSeconds, envAuthzTokensJSON,
		envExternalAuthzURL, envExternalAuthzAudience, envExternalAuthzTimeout,
		envReviewAccountTokenHash, envMetricsLogInterval, envPprofAddr,
		envTracingEnabled, envTracingSampleRatio, envTracingOTLPEndpoint,
		envTracingServiceName, envWebAllowedOrigins, envAssignmentActiveLease,
		envAIAgentClientDev, envAIAgentClientTable, envAIAgentClientSnapshotReload,
		envAIAgentClientHeartbeatSave, envDynamoDBOutboxTable, envAWSRegion,
		envDynamoDBEndpoint, envAgentProfileThumbnailBucket, envAgentProfileThumbnailPrefix,
		envAgentProfileThumbnailCDNBase, envAgentProfileThumbnailMaxBytes,
		envAgentProfileThumbnailExpires, envAgentProfileThumbnailS3Endpoint,
		envTaskContextBaseURL, envTaskContextWorkspaceID, envTaskContextTeamID,
		envTaskContextAPIKey, envTaskContextTimeout, envLongPollMaxHoldSeconds,
		envLongPollTickSeconds, envExternalAuthzAPIKey,
		envAWSContainerCredentialsFullURI, envAWSContainerCredentialsRelativeURI,
		envAWSContainerAuthorizationToken,
	}
}
