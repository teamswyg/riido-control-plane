package riidoaiserver

func dynamoDBStringValue(item map[string]map[string]string, key string) string {
	if item == nil {
		return ""
	}
	return item[key]["S"]
}

func dynamoDBNumberValue(item map[string]map[string]string, key string) string {
	if item == nil {
		return ""
	}
	return item[key]["N"]
}
