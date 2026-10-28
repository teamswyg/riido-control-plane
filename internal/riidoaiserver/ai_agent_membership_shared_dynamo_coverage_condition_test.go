package riidoaiserver

func membershipCoverageCASMatches(p *membershipDynamoAction, previous map[string]map[string]string) bool {
	for _, field := range []struct{ name, parameter, kind string }{
		{"policy_schema", ":schema", "S"},
		{"generation", ":generation", "S"},
		{"credential_generation", ":credential", "S"},
		{"owner_principal_id", ":owner", "S"},
		{"coverage_state", ":state", "S"},
		{"revision", ":revision", "N"},
		{"unresolved", ":unresolved", "N"},
	} {
		parameter, actual := p.ExpressionAttributeValues[field.parameter], previous[field.name]
		expected, parameterOK := parameter[field.kind]
		value, actualOK := actual[field.kind]
		if len(parameter) != 1 || len(actual) != 1 || !parameterOK || !actualOK || value != expected {
			return false
		}
	}
	return true
}
