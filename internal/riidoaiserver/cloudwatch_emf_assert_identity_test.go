package riidoaiserver

import "testing"

func assertCloudWatchEMFIdentity(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	if envelope.AWS.Timestamp != 123456 || envelope.AWS.CloudWatchMetrics[0].Namespace != defaultCloudWatchNamespace {
		t.Fatalf("emf metadata = %+v", envelope.AWS)
	}
	if envelope.AWS.CloudWatchMetrics[0].Dimensions[0][0] != "service" ||
		len(envelope.AWS.CloudWatchMetrics[0].Metrics) == 0 {
		t.Fatalf("emf metric specs = %+v", envelope.AWS.CloudWatchMetrics[0])
	}
	if envelope.SchemaVersion != MetricsSchemaVersion || envelope.Service != defaultCloudWatchServiceName {
		t.Fatalf("emf identity = %+v", envelope)
	}
}

func assertCloudWatchEMFScopes(t *testing.T, envelope cloudWatchEMFEnvelope) {
	t.Helper()
	if envelope.MetricScopeSchemaVersion != cloudWatchEMFMetricScopeSchemaVersion ||
		envelope.StoreStateMetricScope != cloudWatchEMFStoreStateScope ||
		envelope.HTTPMetricScope != cloudWatchEMFRollingWindowScope ||
		envelope.SSEStreamActivityMetricScope != cloudWatchEMFRollingWindowScope ||
		envelope.SSEStreamActiveMetricScope != cloudWatchEMFProcessCurrentScope ||
		envelope.StoreOperationMetricScope != cloudWatchEMFRollingWindowScope ||
		envelope.SnapshotPersistenceMetricScope != cloudWatchEMFProcessLifetimeScope {
		t.Fatalf("emf metric scopes = %+v", envelope)
	}
}
