package riidoaiserver

type dynamoDBAIAgentClientSnapshotWritePlan struct {
	manifest map[string]map[string]string
	parts    []map[string]map[string]string
}

func dynamoDBAIAgentClientSnapshotPlanWrites(items []map[string]map[string]string) dynamoDBAIAgentClientSnapshotWritePlan {
	plan := dynamoDBAIAgentClientSnapshotWritePlan{
		parts: make([]map[string]map[string]string, 0, len(items)),
	}
	for _, item := range items {
		if dynamoDBStringValue(item, "sk") == dynamoDBAIAgentClientSnapshotSK {
			plan.manifest = item
			continue
		}
		plan.parts = append(plan.parts, item)
	}
	return plan
}

func dynamoDBAIAgentClientSnapshotWritePressure(items []map[string]map[string]string) dynamoDBAIAgentClientSnapshotWriteStats {
	plan := dynamoDBAIAgentClientSnapshotPlanWrites(items)
	partsWritten := len(plan.parts)
	return dynamoDBAIAgentClientSnapshotWriteStats{
		itemsWritten: len(items),
		partsSkipped: len(dynamoDBAIAgentClientSnapshotPartNames) - partsWritten,
		partsWritten: partsWritten,
	}
}
