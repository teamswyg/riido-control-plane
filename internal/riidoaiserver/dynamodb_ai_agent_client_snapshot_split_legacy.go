package riidoaiserver

func dynamoDBAIAgentClientSnapshotItemIsLegacy(item map[string]map[string]string) bool {
	return dynamoDBStringValue(item, "snapshot_gzip") != "" ||
		dynamoDBStringValue(item, "snapshot_json") != ""
}

func dynamoDBAIAgentClientSnapshotPartSK(name string) string {
	return "PART#" + name
}
