package riidoaiserver

func buildCloudWatchEMFEnvelope(config CloudWatchEMFConfig, snapshot MetricsSnapshot) cloudWatchEMFEnvelope {
	envelope := cloudWatchEMFEnvelope{
		AWS: cloudWatchEMFMetadata{
			Timestamp: snapshot.GeneratedAt.UnixMilli(),
			CloudWatchMetrics: []cloudWatchMetricGroup{{
				Namespace:  config.Namespace,
				Dimensions: [][]string{{"service"}},
				Metrics:    cloudWatchMetricSpecs(),
			}},
		},
		SchemaVersion:                  snapshot.SchemaVersion,
		Service:                        config.ServiceName,
		MetricScopeSchemaVersion:       cloudWatchEMFMetricScopeSchemaVersion,
		StoreStateMetricScope:          cloudWatchEMFStoreStateScope,
		HTTPMetricScope:                cloudWatchEMFRollingWindowScope,
		StoreOperationMetricScope:      cloudWatchEMFRollingWindowScope,
		SnapshotPersistenceMetricScope: cloudWatchEMFProcessLifetimeScope,
	}
	applyCloudWatchEMFStoreState(&envelope, snapshot)
	applyCloudWatchEMFRuntimeEvents(&envelope, snapshot)
	applyCloudWatchEMFHTTP(&envelope, snapshot)
	applyCloudWatchEMFStoreOperations(&envelope, snapshot)
	applyCloudWatchEMFSnapshotPersistence(&envelope, snapshot)
	return envelope
}
