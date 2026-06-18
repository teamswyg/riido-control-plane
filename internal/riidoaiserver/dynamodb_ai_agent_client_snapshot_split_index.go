package riidoaiserver

import "strings"

func dynamoDBAIAgentClientSnapshotCurrentItem(items []map[string]map[string]string) map[string]map[string]string {
	for _, item := range items {
		if dynamoDBStringValue(item, "sk") == dynamoDBAIAgentClientSnapshotSK {
			return item
		}
	}
	return nil
}

func dynamoDBAIAgentClientSnapshotItemsByPart(items []map[string]map[string]string) map[string]map[string]map[string]string {
	out := make(map[string]map[string]map[string]string, len(items))
	for _, item := range items {
		partName := dynamoDBAIAgentClientSnapshotPartName(item)
		if partName != "" {
			out[partName] = item
		}
	}
	return out
}

func dynamoDBAIAgentClientSnapshotPartHashes(items []map[string]map[string]string) map[string]string {
	out := map[string]string{}
	for _, item := range items {
		partName := dynamoDBStringValue(item, "part_name")
		if partName == "" {
			continue
		}
		if hash := dynamoDBStringValue(item, "part_hash"); hash != "" {
			out[partName] = hash
		}
	}
	return out
}

func dynamoDBAIAgentClientSnapshotPartName(item map[string]map[string]string) string {
	if partName := dynamoDBStringValue(item, "part_name"); partName != "" {
		return partName
	}
	sk := dynamoDBStringValue(item, "sk")
	if !strings.HasPrefix(sk, "PART#") {
		return ""
	}
	return strings.TrimPrefix(sk, "PART#")
}
